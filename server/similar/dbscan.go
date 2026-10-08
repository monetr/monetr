package similar

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/calc"
)

const (
	Epsilon      = 0.5
	MinNeighbors = 1
)

var (
	dbscanClusterDebug = false
)

type Cluster struct {
	Items map[int]uint8
}

type DBSCAN struct {
	// labels is whether we have visited a point yet, by its index in dataset.
	// Every document is one transaction so the index is just as good as the
	// transaction ID here, and a slice lookup is way cheaper than hashing an ID
	// for a map
	labels    []bool
	dataset   []Document
	epsilon   float32
	minPoints int
	clusters  []Cluster
	// scratch is one document's vector expanded back out to the full width, so
	// [calc.SparseDot32] has a dense side to index into. We only allocate it once
	// and reuse it for every point, and it is all zeros whenever we aren't using
	// it
	scratch []float32
}

func NewDBSCAN(dataset []Document, epsilon float32, minPoints int) *DBSCAN {
	// Every document's vector is the same width so one scratch buffer works for
	// all of them
	var scratch []float32
	if len(dataset) > 0 {
		scratch = make([]float32, len(dataset[0].Vector))
	}

	return &DBSCAN{
		labels:    make([]bool, len(dataset)),
		dataset:   dataset,
		epsilon:   epsilon,
		minPoints: minPoints,
		clusters:  nil,
		scratch:   scratch,
	}
}

func (d *DBSCAN) GetDocumentByIndex(index int) (*Document, bool) {
	if index >= len(d.dataset) || index < 0 {
		return nil, false
	}

	return &d.dataset[index], true
}

func (d *DBSCAN) Calculate(ctx context.Context) []Cluster {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Initialize or reinitialize the clusters. We want to start with a clean
	// slate.
	d.clusters = make([]Cluster, 0)
	// From the top, take one point at a time.
	for index := range d.dataset {
		// If we have already visited this point then skip it
		if d.labels[index] {
			continue
		}

		// Find all the other points that are within the epsilon of this point.
		neighbors := d.getNeighbors(index)
		// If there are not enough points then this is not a core point.
		if len(neighbors) < d.minPoints {
			// Mark it as noise and keep moving
			d.labels[index] = true
			continue
		}
		// Otherwise mark the point as visited so we don't do the same work again
		d.labels[index] = true

		// Bootstrap a cluster for the current point
		newCluster := Cluster{
			Items: map[int]uint8{},
		}

		// Then start constructing a cluster around this point.
		d.expandCluster(index, neighbors, &newCluster)
		d.clusters = append(d.clusters, newCluster)
	}

	return d.clusters
}

func (d *DBSCAN) expandCluster(index int, neighbors []int, cluster *Cluster) {
	// And add a pointer to the current item into the new cluster.
	cluster.Items[index] = 0
	for _, neighborIndex := range neighbors {
		// If Q (neighbor) is not visited then mark it as visited and check for more
		// neighbors.
		if !d.labels[neighborIndex] {
			// Mark Q as visited but not as noise.
			d.labels[neighborIndex] = true
			// Find more nearby neighbors.
			newNeighbors := d.getNeighbors(neighborIndex)
			// If we have enough neighbors then we can expand the cluster even more.
			if len(newNeighbors) >= d.minPoints {
				// Merge new neighbors with neighbors. Recursively descend and then add
				// the data we get into the one we currently have.
				d.expandCluster(neighborIndex, newNeighbors, cluster)
			}
		}

		// If Q (neighbor) is not yet part of any cluster
		var found bool
		for _, cluster := range d.clusters {
			_, ok := cluster.Items[neighborIndex]
			if ok {
				found = true
				break
			}
		}
		// Then add it to this cluster.
		if !found {
			cluster.Items[neighborIndex] = 0
		}
	}
}

func (d *DBSCAN) getNeighbors(index int) []int {
	// Pre-allocate an array of neighbors for us to work with.
	neighbors := make([]int, 0)
	point := d.dataset[index]

	// Copy this point into the scratch buffer so SparseDot32 has a dense side to
	// index into. We only write the indicies this point actually has and only
	// clear those again at the end, so this costs however many words are in the
	// transaction instead of the size of the whole vocabulary
	for i, vectorIndex := range point.Indices {
		d.scratch[vectorIndex] = point.Values[i]
	}

	for i, counterpoint := range d.dataset {
		// Don't calculate against yourself
		if i == index {
			continue
		}

		// If two documents don't have a single word in common then they can't be
		// similar. Both vectors are normalized so both squared norms are about 1,
		// and with no words in common the dot product is 0. So the distance would
		// be 1 + 1 - 0 = 2, way past any epsilon we would use
		//
		// The signature has one bit set per word (index % 64), so if no bits
		// overlap then no words overlap. Two different words can land on the same
		// bit, but that just means we do the real calculation below for nothing,
		// it can never make us skip a real neighbor
		if point.Signature&counterpoint.Signature == 0 {
			continue
		}

		// Calculate the distance from our Q point to our P point. This is the same
		// squared distance EuclideanDistance32 gives us, it is just worked out from
		// the dot product (a . b) instead:
		//
		//   ||a - b||^2 = ||a||^2 + ||b||^2 - 2(a . b)
		//
		// We already have both squared norms (Norm2), so the dot product is the
		// only thing left to do for each pair. Every index the counterpoint doesn't
		// have multiplies out to 0, so SparseDot32 only looks at the ones it does
		dot := calc.SparseDot32(d.scratch, counterpoint.Indices, counterpoint.Values)
		distance := point.Norm2 + counterpoint.Norm2 - 2*dot
		// If we are close enough then we could be part of a core cluster point. Add
		// it to the list.
		if distance <= d.epsilon {
			neighbors = append(neighbors, i)
		}
	}

	// Put the scratch buffer back to all zeros for the next point
	for _, vectorIndex := range point.Indices {
		d.scratch[vectorIndex] = 0
	}

	return neighbors
}
