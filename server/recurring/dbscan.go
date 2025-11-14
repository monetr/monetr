package recurring

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
)

const (
	Epsilon      = 0.6
	MinNeighbors = 1
)

var (
	dbscanClusterDebug = false
)

type Cluster struct {
	Items map[int]uint8
}

type DBSCAN struct {
	labels    map[models.ID[models.Transaction]]bool
	dataset   []Document
	epsilon   float32
	minPoints int
	clusters  []Cluster
}

func NewDBSCAN(dataset []Document, epsilon float32, minPoints int) *DBSCAN {
	return &DBSCAN{
		labels:    map[models.ID[models.Transaction]]bool{},
		dataset:   dataset,
		epsilon:   epsilon,
		minPoints: minPoints,
		clusters:  nil,
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

	// Initialize or reinitialize the clusters. We want to start with a clean slate.
	d.clusters = make([]Cluster, 0)
	// From the top, take one point at a time.
	for index, point := range d.dataset {
		// If we have already visited this point then skip it
		if _, visited := d.labels[point.ID]; visited {
			continue
		}

		// Find all the other points that are within the epsilon of this point.
		neighbors := d.getNeighbors(index)
		// If there are not enough points then this is not a core point.
		if len(neighbors) < d.minPoints {
			// Mark it as noise and keep moving
			d.labels[point.ID] = true
			continue
		}
		// Otherwise mark the point as visited so we don't do the same work again
		d.labels[point.ID] = false

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
		// Retrieve the item from the dataset.
		neighbor := d.dataset[neighborIndex]
		// If Q (neighbor) is not visited then mark it as visited and check for more neighbors.
		if _, visited := d.labels[neighbor.ID]; !visited {
			// Mark Q as visited but not as noise.
			d.labels[neighbor.ID] = false
			// Find more nearby neighbors.
			newNeighbors := d.getNeighbors(neighborIndex)
			// If we have enough neighbors then we can expand the cluster even more.
			if len(newNeighbors) >= d.minPoints {
				// Merge new neighbors with neighbors
				// Recursively descend and then add the data we get into the one we currently have.
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
	neighbors := make([]int, 0, len(d.dataset)/2)
	point := d.dataset[index]
	for i, counterpoint := range d.dataset {
		// Don't calculate against yourself
		if i == index {
			continue
		}

		// Calculate the distance from our Q point to our P point.
		distance := calc.EuclideanDistance32(point.Vector, counterpoint.Vector)
		// If we are close enough then we could be part of a core cluster point. Add it to the list.
		if distance <= d.epsilon {
			neighbors = append(neighbors, i)
		}
	}

	return neighbors
}
