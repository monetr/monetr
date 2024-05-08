package recurring

import (
	"math"
	"regexp"

	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
)

var (
	// clusterCleanStringRegex = regexp.MustCompile(`[a-zA-Z'\.\d]+`)
	clusterCleanStringRegex = regexp.MustCompile(`(?:\b(?:[a-zA-Z]|\d){1}(?:[a-zA-Z.]{1,})(?:\d{1}[a-zA-Z]*){0,2}\b)`)
	numberOnly              = regexp.MustCompile(`^\d+$`)

	specialWeights = map[string]float32{
		"amazon":          10,
		"youtube premium": 5,
		"google":          2,
		"pwp":             0,
		"debit":           0,
		"pos":             0,
		"visa":            0,
		"ach":             0,
		"transaction":     0,
		"card":            0,
		"check":           0,
		"transfer":        0,
		"deposit":         0,
		"purchase":        0,
		"adjustment":      0,
		"helppay":         0, // Shows up on some google transactions, not helpful.
	}

	synonyms = map[string]string{
		"amzn":       "amazon",
		"youtubepre": "youtube premium",
	}
)

type Document struct {
	ID          models.ID[models.Transaction]
	TF          map[string]float32
	TFIDF       map[string]float32
	Vector      []float32
	Transaction *models.Transaction
	Parts       []string
	Valid       bool
}

type TFIDF struct {
	documents []Document
	wc        map[string]float32
	idf       map[string]float32
}

func NewTransactionTFIDF() *TFIDF {
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
	lower, normal := CleanNameRegex(txn)
	name := make([]string, 0, len(lower))
	wordCounts := make(map[string]float32, len(lower))
	for i, word := range lower {
		// If there is a synonym for the current word use that instead.
		if synonym, ok := synonyms[word]; ok {
			word = synonym
		}
		if multiplier, ok := specialWeights[word]; ok && multiplier == 0 {
			continue
		}
		wordCounts[word]++
		p.wc[word]++
		name = append(name, normal[i])
	}

	tf := make(map[string]float32, len(wordCounts))
	for word, count := range wordCounts {
		tf[word] = count / float32(len(lower))
	}

	p.documents = append(p.documents, Document{
		ID:          txn.TransactionId,
		Parts:       name,
		Transaction: txn,
		TF:          tf,
		TFIDF:       map[string]float32{},
	})
}

func (p *TFIDF) GetDocuments() []Document {
	resultDocuments := make([]Document, 0, len(p.documents))
	docCount := float32(len(p.documents))
	p.idf = make(map[string]float32, len(p.wc))
	for word, count := range p.wc {
		p.idf[word] = float32(math.Log(float64(docCount / (count + 1))))
	}
	// Get a map of all the meaningful words and their index to use in the vector
	minified := p.indexWords()
	// Define the length of the vector and adjust it to be divisible by 16. This will enable us to leverage SIMD in the
	// future. By using 16 we are compatible with both AVX and AVX512.
	vectorLength := len(minified) + (16 - (len(minified) % 16))
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
			resultDocuments = append(resultDocuments, document)
		}
	}

	return resultDocuments
}
