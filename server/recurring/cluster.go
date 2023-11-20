package recurring

import (
	"math"
	"strings"
)

type Document struct {
	Transaction  *Transaction
	TF           map[string]float64
	TFIDF        map[string]float64
	AmountScaled float64
	Vector       []float64
}

type WordCount struct {
	index int
	wc    map[string][2]int
}

func (w *WordCount) Increment(word string) {
	value, ok := w.wc[word]
	if !ok {
		w.wc[word] = [2]int{
			w.index,
			1,
		}
		w.index++
		return
	}

	value[1]++
	w.wc[word] = value
}

func (w *WordCount) Iterate(callback func(word string, index int, count int)) {
	for word, value := range w.wc {
		callback(word, value[0], value[1])
	}
}

func (w *WordCount) Index(word string) int {
	value, ok := w.wc[word]
	if !ok {
		return -1
	}
	return value[0]
}

func (w *WordCount) Minify() map[string]int {
	index := 0
	result := make(map[string]int)
	for word, value := range w.wc {
		if value[1] == 1 {
			continue
		}
		result[word] = index
		index++
	}

	return result
}

type PreProcessor struct {
	documents []Document
	// Word count
	wc  *WordCount
	idf map[string]float64
}

func (p *PreProcessor) AddTransaction(txn *Transaction) {
	words := cleanStringRegex.FindAllString(txn.OriginalName, len(txn.OriginalName))
	if txn.OriginalMerchantName != nil {
		words = append(words, strings.Fields(*txn.OriginalMerchantName)...)
	}

	wordCounts := make(map[string]int, len(words))
	// Get the term frequency from the transaction name
	for _, word := range words {
		word = strings.ToLower(word)
		wordCounts[word]++
		p.wc.Increment(word)
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
	p.wc.Iterate(func(word string, _ int, count int) {
		p.idf[word] = math.Log(docCount / (float64(count) + 1))
	})
	var min, max float64
	for i := range p.documents {
		document := p.documents[i]
		for word, tfValue := range document.TF {
			document.TFIDF[word] = tfValue * p.idf[word]
		}
		p.documents[i] = document
		amount := float64(document.Transaction.Amount)
		if amount > max || max == 0 {
			max = amount
		}
		if amount < min || min == 0 {
			min = amount
		}
	}

	minified := p.wc.Minify()
	for i := range p.documents {
		document := p.documents[i]
		amount := float64(document.Transaction.Amount)
		adjusted := (amount - min) / (max - min)
		document.AmountScaled = 2*adjusted - 1

		// Calculate a vector based on all the words from all transactions.
		// TODO Also need to look into Principal Component Analysis
		vector := make([]float64, len(minified))
		for word, tfidfValue := range document.TFIDF {
			index, exists := minified[word]
			if !exists {
				continue
			}
			vector[index] = tfidfValue
		}
		var norm float64
		for _, value := range vector {
			norm += value * value
		}
		norm = math.Sqrt(norm)
		for i, value := range vector {
			vector[i] = value / norm
		}
		document.Vector = append([]float64{document.AmountScaled}, vector...)
		p.documents[i] = document
	}
}

func (p *PreProcessor) GetDatums() []Datum {
	datums := make([]Datum, len(p.documents))
	for i, document := range p.documents {
		datums[i] = Datum{
			ID:          document.Transaction.TransactionId,
			Transaction: *document.Transaction,
			Vector:      document.Vector,
		}
	}

	return datums
}

type Datum struct {
	ID          uint64
	Transaction Transaction
	Vector      []float64
}

func (a Datum) Distance(b Datum) float64 {
	var distance float64
	for i, value := range a.Vector {
		distance += math.Pow(value-b.Vector[i], 2)
	}
	return distance
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
	d.clusters = make([]Cluster, 0)
	for index, point := range d.dataset {
		// If we have already visited this point then skip it
		if _, visited := d.labels[point.ID]; visited {
			continue
		}

		neighbors := d.getNeighbors(index)
		if len(neighbors) < d.minPoints {
			d.labels[point.ID] = true
			continue
		}
		// mark point as visited
		d.labels[point.ID] = false
		d.clusters = append(d.clusters, d.expandCluster(index, neighbors))
	}

	return d.clusters
}

func (d *DBSCAN) expandCluster(index int, neighbors []int) Cluster {
	cluster := Cluster{
		Items: map[int]Datum{},
	}
	cluster.Items[index] = d.dataset[index]
	for _, neighborIndex := range neighbors {
		neighbor := d.dataset[neighborIndex]
		// IF Q is not visited
		if _, visited := d.labels[neighbor.ID]; !visited {
			// Mark Q as visited
			d.labels[neighbor.ID] = false
			newNeighbors := d.getNeighbors(neighborIndex)
			if len(newNeighbors) >= d.minPoints {
				// Merge new neighbors with neighbors
				// TODO Very evil to modify an array while inside it!
				neighbors = append(neighbors, newNeighbors...)
			}
		}

		// if Q is not yet part of any cluster
		var found bool
		for _, cluster := range d.clusters {
			_, ok := cluster.Items[neighborIndex]
			if ok {
				found = true
				break
			}
		}
		if !found {
			cluster.Items[neighborIndex] = neighbor
		}
	}

	return cluster
}

func (d *DBSCAN) getNeighbors(index int) []int {
	neighbors := make([]int, 0, len(d.dataset))
	point := d.dataset[index]
	for i, counterpoint := range d.dataset {
		// Don't calculate against yourself
		if i == index {
			continue
		}

		distance := point.Distance(counterpoint)
		if distance <= d.epsilon {
			neighbors = append(neighbors, i)
		}
	}

	return neighbors
}
