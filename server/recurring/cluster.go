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

	specialWeights = map[string]float64{
		"amazon":      10,
		"pwp":         0,
		"debit":       0,
		"pos":         0,
		"visa":        0,
		"ach":         0,
		"transaction": 0,
		"card":        0,
		"check":       0,
	}

	synonyms = map[string]string{
		"amzn": "amazon",
	}
)

type Document struct {
	ID          uint64
	TF          map[string]float64
	TFIDF       map[string]float64
	Vector      []float64
	Transaction *models.Transaction
	String      string
	Valid       bool
}

type PreProcessor struct {
	documents []Document
	// Word count
	wc  map[string]int
	idf map[string]float64
}

func (p *PreProcessor) indexWords() map[string]int {
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

func (p *PreProcessor) AddTransaction(txn *models.Transaction) {
	words := clusterCleanStringRegex.FindAllString(txn.OriginalName, len(txn.OriginalName))
	if txn.OriginalMerchantName != "" {
		words = append(words, clusterCleanStringRegex.FindAllString(txn.OriginalMerchantName, len(txn.OriginalMerchantName))...)
	}

	name := make([]string, 0, len(words))
	wordCounts := make(map[string]int, len(words))
	// Get the term frequency from the transaction name
	for _, word := range words {
		word = strings.ToLower(word)
		word = strings.ReplaceAll(word, "'", "")
		word = strings.ReplaceAll(word, ".", "")

		numbers := numberOnly.FindAllString(word, len(word))
		if len(numbers) > 0 {
			continue
		}

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

	tf := make(map[string]float64, len(wordCounts))
	for word, count := range wordCounts {
		tf[word] = float64(count) / float64(len(words))
	}

	p.documents = append(p.documents, Document{
		ID:          txn.TransactionId,
		String:      strings.Join(name, " "),
		Transaction: txn,
		TF:          tf,
		TFIDF:       map[string]float64{},
	})
}

// PostPrepareCalculations should be called after all transactions have been added to the dataset.
func (p *PreProcessor) PostPrepareCalculations() {
	docCount := float64(len(p.documents))
	for word, count := range p.wc {
		p.idf[word] = math.Log(docCount / (float64(count) + 1))
	}
	// Get a map of all the meaningful words and their index to use in the vector
	minified := p.indexWords()
	// Define the length of the vector and adjust it to be divisible by 8. This will enable us to leverage SIMD in the
	// future. By using 8 we are compatible with both AVX and AVX512.
	vectorLength := len(minified) + (8 - (len(minified) % 8))
	for i := range p.documents {
		// Get the current document we are working with
		document := p.documents[i]
		// Calculate the TFIDF for that document
		for word, tfValue := range document.TF {
			document.TFIDF[word] = tfValue * p.idf[word]
			// If this specific word is meant to be more meaningful than tfidf might treat it then adjust it accordingly
			if multiplier, ok := specialWeights[word]; ok {
				document.TFIDF[word] *= multiplier
			}
		}
		// Then create a vector of the words in the document name to use for the DBSCAN clustering
		document.Vector = make([]float64, vectorLength)
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
		calc.NormalizeVector64(document.Vector)
		// Then store the document back in
		p.documents[i] = document
	}
}

func (p *PreProcessor) GetDatums() []Datum {
	datums := make([]Datum, 0, len(p.documents))
	for _, document := range p.documents {
		if !document.Valid {
			continue
		}
		datums = append(datums, Datum{
			ID:          document.ID,
			Transaction: document.Transaction,
			String:      document.String,
			Amount:      document.Transaction.Amount,
			Vector:      document.Vector,
		})
	}

	return datums
}

type Datum struct {
	ID          uint64
	Transaction *models.Transaction
	String      string
	Amount      int64
	Vector      []float64
}

type Cluster struct {
	ID    uint64
	Items map[int]uint8
}

type DBSCAN struct {
	labels    map[uint64]bool
	dataset   []Datum
	epsilon   float64
	minPoints int
	clusters  []Cluster
}

func NewDBSCAN(dataset []Datum, epsilon float64, minPoints int) *DBSCAN {
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
		distance := calc.EuclideanDistance64(point.Vector, counterpoint.Vector)
		// If we are close enough then we could be part of a core cluster point. Add it to the list.
		if distance <= d.epsilon {
			neighbors = append(neighbors, i)
		}
	}

	return neighbors
}
