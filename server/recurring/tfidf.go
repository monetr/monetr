package recurring

import (
	"context"
	"math"
	"regexp"
	"strings"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
)

var (
	clusterCleanStringRegex = regexp.MustCompile(`(?:\b(?:[a-zA-Z]|\d){1}(?:[a-zA-Z.']{1,})(?:\d{1}[a-zA-Z]*){0,2}\b)`)
	vowelsOnly              = regexp.MustCompile(`[aeyiuo]+`)
	numberOnly              = regexp.MustCompile(`^\d+$`)

	specialWeights = map[string]float32{
		"amazon":          10,
		"youtube premium": 5,
		"google":          2,
		"stripe":          2,
		"pwp":             0, // Paid with privacy
		"privacycom":      0, // Same as pwp
		"sq":              0, // Square
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
		"payment":         0,
		"to":              0,
		"of":              0,
		"helppay":         0, // Shows up on some google transactions, not helpful.
		"null":            0, // Shows up in manual imports somtimes

		"merchant": 0, // Shows up in almost all mercury transactions.
		"name":     0, // Shows up in almost all mercury transactions.

		// TODO, get a list of country codes to exclude?
		"us": 0,
	}

	states = map[string]float32{
		"al": 0,
		"ak": 0,
		"az": 0,
		"ar": 0,
		"ca": 0,
		"co": 0,
		"ct": 0,
		"de": 0,
		"fl": 0,
		"ga": 0,
		"hi": 0,
		"id": 0,
		"il": 0,
		"in": 0,
		"ia": 0,
		"ks": 0,
		"ky": 0,
		"la": 0,
		"me": 0,
		"md": 0,
		"ma": 0,
		"mi": 0,
		"mn": 0,
		"ms": 0,
		"mo": 0,
		"mt": 0,
		"ne": 0,
		"nv": 0,
		"nh": 0,
		"nj": 0,
		"nm": 0,
		"ny": 0,
		"nc": 0,
		"nd": 0,
		"oh": 0,
		"ok": 0,
		"or": 0,
		"pa": 0,
		"ri": 0,
		"sc": 0,
		"sd": 0,
		"tn": 0,
		"tx": 0,
		"ut": 0,
		"vt": 0,
		"va": 0,
		"wa": 0,
		"wv": 0,
		"wi": 0,
		"wy": 0,
	}

	synonyms = map[string]string{
		"amzn":        "amazon",
		"youtubepre":  "youtube premium",
		"youtubeprem": "youtube premium",
		"coffe":       "coffee",
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

func TokenizeName(txn *models.Transaction) (lower, normal []string) {
	lowerIn, normalIn := CleanNameRegex(txn)
	lower = make([]string, 0, len(lowerIn))
	normal = make([]string, 0, len(lowerIn))
	for i, word := range lowerIn {
		item := normalIn[i]
		// If there is a synonym for the current word use that instead.
		if synonym, ok := synonyms[word]; ok {
			item = synonym
		}
		if multiplier, ok := specialWeights[word]; ok && multiplier == 0 {
			continue
		}
		if _, ok := states[word]; ok {
			// Exclude states from names
			continue
		}
		lower = append(lower, strings.ToLower(item))
		normal = append(normal, item)
	}
	return lower, normal
}

func (p *TFIDF) AddTransaction(txn *models.Transaction) {
	lower, upper := TokenizeName(txn)
	wordCounts := make(map[string]float32, len(lower))
	for _, word := range lower {
		wordCounts[word]++
		p.wc[word]++
	}

	tf := make(map[string]float32, len(wordCounts))
	for word, count := range wordCounts {
		tf[word] = count / float32(len(lower))
	}

	p.documents = append(p.documents, Document{
		ID:          txn.TransactionId,
		Parts:       upper,
		Transaction: txn,
		TF:          tf,
		TFIDF:       map[string]float32{},
	})
}

func (p *TFIDF) GetDocuments(ctx context.Context) []Document {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	crumbs.Debug(span.Context(), "Organizing documents for DBSCAN clustering (transaction similarity)", map[string]any{
		"count": len(p.documents),
	})

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
