package recurring

import (
	"math"
	"sort"
	"strings"
)

var (
	specialWeights = map[string]float64{
		"amazon": 10,
	}

	synonyms = map[string]string{
		"amzn": "amazon",
	}
)

type Document struct {
	Transaction  *Transaction
	TF           map[string]float64
	TFIDF        map[string]float64
	AmountScaled float64
	Vector       []float64
	Valid        bool
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

func (p *PreProcessor) AddTransaction(txn *Transaction) {
	words := cleanStringRegex.FindAllString(txn.OriginalName, len(txn.OriginalName))
	if txn.OriginalMerchantName != nil {
		words = append(words, cleanStringRegex.FindAllString(*txn.OriginalMerchantName, len(*txn.OriginalMerchantName))...)
	}

	wordCounts := make(map[string]int, len(words))
	// Get the term frequency from the transaction name
	for _, word := range words {
		word = strings.ToLower(word)
		// If there is a synonym for the current word use that instead.
		if synonym, ok := synonyms[word]; ok {
			word = synonym
		}
		wordCounts[word]++
		p.wc[word]++
	}

	tf := make(map[string]float64, len(wordCounts))
	for word, count := range wordCounts {
		tf[word] = float64(count) / float64(len(words))
	}

	p.documents = append(p.documents, Document{
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
	// Define the length of the vector and adjust it to be divisible by 4. This will enable us to leverage SIMD in the
	// future.
	vectorLength := len(minified) + (len(minified) % 4)
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
		var norm float64
		for _, value := range document.Vector {
			norm += value * value
		}
		norm = math.Sqrt(norm)
		for i, value := range document.Vector {
			document.Vector[i] = value / norm
		}
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
			ID:          document.Transaction.TransactionId,
			Transaction: *document.Transaction,
			Vector:      document.Vector,
		})
	}

	return datums
}

type Datum struct {
	ID          uint64
	Transaction Transaction
	Vector      []float64
}

func (a Datum) Distance(b Datum) float64 {
	// This is just the Euclidean distance between two points since the vectors here have many many dimensions.
	var distance float64
	for i, value := range a.Vector {
		distance += math.Pow(value-b.Vector[i], 2)
	}
	return distance
}

func kDistances(data []Datum, minPoints int) []float64 {
	distances := make([]float64, 0, len(data))
	for _, p := range data {
		pointDistances := make([]float64, 0, len(data))
		for _, q := range data {
			// A point should not be able to see itself.
			if q.ID == p.ID {
				continue
			}

			// Add the distance from p -> q to the current dataset
			pointDistances = append(pointDistances, p.Distance(q))
		}

		// If we have not accumulated enough data points for p then it isn't good enough to be part of the set. Throw it out
		// and keep going.
		if len(pointDistances) < minPoints {
			continue
		}
		sort.Float64s(pointDistances)
		// Take the distance that is to the most extreme.
		// TODO Double check this is correct, I can never remember if sort is ascending or descending.
		distances = append(distances, pointDistances[minPoints-1])
	}
	// Sort it again before returning it to the caller
	sort.Float64s(distances)
	return distances
}

type Cluster struct {
	Items map[int]Datum
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
		d.clusters = append(d.clusters, d.expandCluster(index, neighbors))
	}

	return d.clusters
}

func (d *DBSCAN) expandCluster(index int, neighbors []int) Cluster {
	// Bootstrap a cluster for the current point, this function might be called recursively
	cluster := Cluster{
		Items: map[int]Datum{},
	}
	// And add a pointer to the current item into the new cluster.
	cluster.Items[index] = d.dataset[index]
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
			cluster.Items[neighborIndex] = neighbor
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
		distance := point.Distance(counterpoint)
		// If we are close enough then we could be part of a core cluster point. Add it to the list.
		if distance <= d.epsilon {
			neighbors = append(neighbors, i)
		}
	}

	return neighbors
}
