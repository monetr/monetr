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
	// labels records whether a point has been visited yet, indexed the same way
	// as dataset. This used to be a map keyed by the transaction ID, but it gets
	// probed once for every pair of points, and hashing an ID that many times
	// ended up costing several times more than the distance calculation it was
	// guarding. Each document owns exactly one transaction ID so indexing by
	// position is equivalent.
	labels    []bool
	dataset   []Document
	epsilon   float32
	minPoints int
	clusters  []Cluster
	// scratch is where a single document gets expanded back into a dense vector
	// so that the sparse distance kernel has something to index into. It is
	// allocated once and reused for every point, and is always all zeros in
	// between uses.
	scratch []float32
}

func NewDBSCAN(dataset []Document, epsilon float32, minPoints int) *DBSCAN {
	// Every document's vector is the same width, so a single scratch buffer that
	// size can serve all of them.
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

	// Expand this point into the scratch buffer so the sparse distance kernel has
	// a dense side to index into. Only the indicies this document occupies get
	// written, and they are the only ones cleared again at the end, so the cost
	// of this is the number of words in the transaction rather than the size of
	// the whole vocabulary.
	for i, vectorIndex := range point.Indices {
		d.scratch[vectorIndex] = point.Values[i]
	}

	for i, counterpoint := range d.dataset {
		// Don't calculate against yourself
		if i == index {
			continue
		}

		// Two documents that do not have a single word in common cannot be
		// similar. Both vectors are normalized, so the distance between them would
		// come out at roughly 2.0, which is far beyond any epsilon worth using.
		// The signature is a bloom filter of the word indicies, so no overlapping
		// bits proves there are no overlapping words. A collision can only produce
		// a false positive, and that just falls through to the real calculation
		// below, so this can never drop a neighbor that should have been kept.
		if point.Signature&counterpoint.Signature == 0 {
			continue
		}

		// Calculate the distance from our Q point to our P point. The dot product
		// only needs the indicies the counterpoint occupies, because every other
		// index multiplies out to zero, and the squared distance falls out of it:
		// ||a - b||^2 == ||a||^2 + ||b||^2 - 2(a . b)
		dot := calc.SparseDot32(d.scratch, counterpoint.Indices, counterpoint.Values)
		distance := point.Norm2 + counterpoint.Norm2 - 2*dot
		// If we are close enough then we could be part of a core cluster point. Add
		// it to the list.
		if distance <= d.epsilon {
			neighbors = append(neighbors, i)
		}
	}

	// Put the scratch buffer back the way we found it for the next point.
	for _, vectorIndex := range point.Indices {
		d.scratch[vectorIndex] = 0
	}

	return neighbors
}
