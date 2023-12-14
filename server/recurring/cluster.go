package recurring

import (
	"math"
	"regexp"
	"strings"

	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
)

const (
	Epsilon      = 0.98
	MinNeighbors = 1
)

var (
	dbscanClusterDebug = false
)

var (
	clusterCleanStringRegex = regexp.MustCompile(`[a-zA-Z'\.\d]+`)
	numberOnly              = regexp.MustCompile(`^\d+$`)

	specialWeights = map[string]float32{
		"amazon":      10,
		"pwp":         0,
		"debit":       0,
		"pos":         0,
		"visa":        0,
		"ach":         0,
		"transaction": 0,
		"card":        0,
		"check":       0,
		"transfer":    0,
		"deposit":     0,
	}

	synonyms = map[string]string{
		"amzn": "amazon",
	}
)

type Document struct {
	ID          uint64
	TF          map[string]float32
	TFIDF       map[string]float32
	Vector      []float32
	Transaction *models.Transaction
	String      string
	Valid       bool
}

type TFIDF struct {
	documents []Document
	wc        map[string]float32
}

func NewPreProcessor() *TFIDF {
	return &TFIDF{
		documents: []Document{},
		wc:        map[string]float32{},
	}
}

func (p *TFIDF) indexWords() map[string]int {
	index := 0
	result := make(map[string]int)
	for word, count := range p.wc {
		if count == 1 {
			continue
		}
		result[word] = index
		index++
	}

	return result
}

func (p *TFIDF) AddTransaction(txn *models.Transaction) {
	words := CleanNameRegex(txn)
	name := make([]string, 0, len(words))
	wordCounts := make(map[string]float32, len(words))
	for _, word := range words {
		// If there is a synonym for the current word use that instead.
		if synonym, ok := synonyms[word]; ok {
			word = synonym
		}
		if multiplier, ok := specialWeights[word]; ok && multiplier == 0 {
			continue
		}
		wordCounts[word]++
		p.wc[word]++
		name = append(name, word)
	}

	tf := make(map[string]float32, len(wordCounts))
	for word, count := range wordCounts {
		tf[word] = count / float32(len(words))
	}

	p.documents = append(p.documents, Document{
		ID:          txn.TransactionId,
		String:      strings.Join(name, " "),
		Transaction: txn,
		TF:          tf,
		TFIDF:       map[string]float32{},
	})
}

func (p *TFIDF) GetDatums() []Datum {
	datums := make([]Datum, 0, len(p.documents))
	docCount := float32(len(p.documents))
	idf := make(map[string]float32, len(p.wc))
	for word, count := range p.wc {
		idf[word] = float32(math.Log(float64(docCount / (count + 1))))
	}
	// Get a map of all the meaningful words and their index to use in the vector
	minified := p.indexWords()
	// Define the length of the vector and adjust it to be divisible by 8. This will enable us to leverage SIMD in the
	// future. By using 8 we are compatible with both AVX and AVX512.
	vectorLength := len(minified) + (16 - (len(minified) % 16))
	for i := range p.documents {
		// Get the current document we are working with
		document := p.documents[i]
		// Calculate the TFIDF for that document
		for word, tfValue := range document.TF {
			document.TFIDF[word] = tfValue * idf[word]
			// If this specific word is meant to be more meaningful than tfidf might treat it then adjust it accordingly
			if multiplier, ok := specialWeights[word]; ok {
				document.TFIDF[word] *= multiplier
			}
		}
		// Then create a vector of the words in the document name to use for the DBSCAN clustering
		document.Vector = make([]float32, vectorLength)
		words := 0
		for word, tfidfValue := range document.TFIDF {
			index, exists := minified[word]
			if !exists {
				continue
			}
			words++
			document.Vector[index] = tfidfValue
		}
		if words == 0 {
			document.Valid = false
			p.documents[i] = document
			continue
		}
		document.Valid = true
		// Normalize the document's tfidf vector.
		calc.NormalizeVector32(document.Vector)
		p.documents[i] = document
		// Then store the document back in
		if document.Valid {
			datums = append(datums, Datum{
				ID:          document.ID,
				Transaction: document.Transaction,
				String:      document.String,
				Amount:      document.Transaction.Amount,
				Vector:      document.Vector,
			})
		}
	}

	return datums
}

type Datum struct {
	ID          uint64
	Transaction *models.Transaction
	String      string
	Amount      int64
	Vector      []float32
}

type Cluster struct {
	ID    uint64
	Items map[int]uint8
}

type DBSCAN struct {
	labels    map[uint64]bool
	dataset   []Datum
	epsilon   float32
	minPoints int
	clusters  []Cluster
}

func NewDBSCAN(dataset []Datum, epsilon float32, minPoints int) *DBSCAN {
	return &DBSCAN{
		labels:    map[uint64]bool{},
		dataset:   dataset,
		epsilon:   epsilon,
		minPoints: minPoints,
		clusters:  nil,
	}
}

func (d *DBSCAN) GetDatumByIndex(index int) (*Datum, bool) {
	if index >= len(d.dataset) || index < 0 {
		return nil, false
	}

	return &d.dataset[index], true
}

func (d *DBSCAN) Calculate() []Cluster {
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
		// Then start constructing a cluster around this point.
		newCluster := d.expandCluster(index, neighbors)
		// Set the cluster's unique ID to the lowest numeric ID in that cluster.
		// HACK: I need a way to uniquely identify each cluster. Generally by using the contents of that cluster. This
		// relies on the contents of that cluster remaining consistent over time. While the order of the clusters might
		// change in the future or they might expand as new transactions show up, I need to know which cluster they get
		// added to in order to tune things over time. This has the potential to cause issues on its own, what if the
		// cluster algorithm changes enough that the "lowest ID" gets kicked out of the cluster? What if we push a bad
		// change and the clusters change entirely? Or what if that "lowest ID" gets moved to a different cluster. This
		// needs improvement, but I think this should be fine for the initial implementation of the clustering algorithm.
		for i := range newCluster.Items {
			item := d.dataset[i]
			if item.ID < newCluster.ID || newCluster.ID == 0 {
				newCluster.ID = item.ID
			}
		}

		d.clusters = append(d.clusters, newCluster)
	}

	return d.clusters
}

func (d *DBSCAN) expandCluster(index int, neighbors []int) Cluster {
	// Bootstrap a cluster for the current point, this function might be called recursively
	cluster := Cluster{
		Items: map[int]uint8{},
	}
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
				recurResult := d.expandCluster(neighborIndex, newNeighbors)
				// Just add the recurred items into this cluster.
				for k, v := range recurResult.Items {
					cluster.Items[k] = v
				}
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

	return cluster
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
