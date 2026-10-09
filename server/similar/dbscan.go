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
	// [calc.SparseNeighbors32] has a dense side to index into. We only allocate
	// it once and reuse it for every point, and it is all zeros whenever we
	// aren't using it
	scratch []float32
	// These are the sparse vectors of every document laid out end to end, the
	// way [calc.SparseNeighbors32] wants them. Document i's entries are
	// indices[offsets[i]:offsets[i+1]] and values[offsets[i]:offsets[i+1]].
	// Looking at the next document is just the next few int32s instead of
	// jumping over a whole Document struct, which is way easier on the cache
	signatures []uint64
	norms      []float32
	offsets    []int32
	indices    []int32
	values     []float32
	// neighbors is where [calc.SparseNeighbors32] writes the neighbors of a
	// point. Every document could be a neighbor so it is as long as the dataset
	neighbors []int32
}

func NewDBSCAN(dataset []Document, epsilon float32, minPoints int) *DBSCAN {
	// Every document's vector is the same width so one scratch buffer works for
	// all of them
	var scratch []float32
	if len(dataset) > 0 {
		scratch = make([]float32, len(dataset[0].Vector))
	}

	var entries int
	for i := range dataset {
		entries += len(dataset[i].Indices)
	}

	signatures := make([]uint64, len(dataset))
	norms := make([]float32, len(dataset))
	offsets := make([]int32, len(dataset)+1)
	indices := make([]int32, 0, entries)
	values := make([]float32, 0, entries)
	for i := range dataset {
		signatures[i] = dataset[i].Signature
		norms[i] = dataset[i].Norm2
		indices = append(indices, dataset[i].Indices...)
		values = append(values, dataset[i].Values...)
		offsets[i+1] = int32(len(indices))
	}

	return &DBSCAN{
		labels:     make([]bool, len(dataset)),
		dataset:    dataset,
		epsilon:    epsilon,
		minPoints:  minPoints,
		clusters:   nil,
		scratch:    scratch,
		signatures: signatures,
		norms:      norms,
		offsets:    offsets,
		indices:    indices,
		values:     values,
		neighbors:  make([]int32, len(dataset)),
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
	point := d.dataset[index]

	// Copy this point into the scratch buffer so SparseNeighbors32 has a dense
	// side to index into. We only write the indicies this point actually has and
	// only clear those again at the end, so this costs however many words are in
	// the transaction instead of the size of the whole vocabulary
	for i, vectorIndex := range point.Indices {
		d.scratch[vectorIndex] = point.Values[i]
	}

	// This checks every document against our point in one call.
	//
	// If two documents don't have a single word in common then they can't be
	// similar. Both vectors are normalized so both squared norms are about 1, and
	// with no words in common the dot product is 0. So the distance would be
	// 1 + 1 - 0 = 2, way past any epsilon we would use. The signature has one bit
	// set per word (index % 64), so if no bits overlap then no words overlap and
	// that document gets skipped. Two different words can land on the same bit,
	// but that just means we do the real calculation for nothing, it can never
	// make us skip a real neighbor
	//
	// For the rest it works out the same squared distance EuclideanDistance32
	// gives us, just from the dot product (a . b) instead:
	//
	//   ||a - b||^2 = ||a||^2 + ||b||^2 - 2(a . b)
	//
	// Anything within epsilon could be part of a core cluster point, and gets
	// written into d.neighbors
	count := calc.SparseNeighbors32(
		d.scratch,
		point.Signature,
		point.Norm2,
		d.epsilon,
		d.signatures,
		d.norms,
		d.offsets,
		d.indices,
		d.values,
		d.neighbors,
	)

	// Put the scratch buffer back to all zeros for the next point
	for _, vectorIndex := range point.Indices {
		d.scratch[vectorIndex] = 0
	}

	// d.neighbors gets written over by the next call, and expandCluster still
	// holds onto these while it calls us again. So they have to be copied out
	neighbors := make([]int, 0, count)
	for _, neighbor := range d.neighbors[:count] {
		// Our point is always going to be close to itself, don't include it
		if int(neighbor) == index {
			continue
		}
		neighbors = append(neighbors, int(neighbor))
	}

	return neighbors
}
