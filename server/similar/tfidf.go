package similar

import (
	"context"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
)

var (
	clusterCleanStringRegex = regexp.MustCompile(`(?:\b(?:[a-zA-Z]|\d){1}(?:[a-zA-Z.']{1,})(?:\d{1}[a-zA-Z]*){0,2}\b)|&`)
	vowelsOnly              = regexp.MustCompile(`[aeyiuo]+`)
	numberOnly              = regexp.MustCompile(`^\d+$`)

	specialWeights = map[string]float32{
		"null": 0, // Shows up in manual imports somtimes

		"inc": 0, // Common busines suffix, not valauble

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
		"amz":         "amazon",
		"amzn":        "amazon",
		"amzncom":     "amazon",
		"amazoncom":   "amazon",
		"youtubepre":  "youtube premium",
		"youtubeprem": "youtube premium",
		"coffe":       "coffee",
	}
)

type Document struct {
	ID     models.ID[models.Transaction]
	TF     map[string]float32
	TFIDF  map[string]float32
	Vector []float32
	// Indices are the positions within Vector that this document actually
	// occupies, in ascending order, and Values are the values sitting at each of
	// those positions. Together they are the sparse form of Vector. A transaction
	// name only contains a handful of meaningful words but the vector is as wide
	// as the entire vocabulary of the account, so the overwhelming majority of
	// Vector is zeros that the distance calculation does not need to look at.
	Indices []int32
	Values  []float32
	// Norm2 is the squared magnitude of the vector. Vector is normalized so this
	// lands very close to 1, but not exactly, and the sparse distance identity
	// wants the real value rather than an assumed one.
	Norm2 float32
	// Signature is a 64 bit bloom filter of Indices. Two documents with no bits
	// in common share no words at all, which means they cannot possibly be
	// similar, and the distance between them never needs to be calculated.
	Signature   uint64
	Tokens      []Token
	Transaction *models.Transaction
	Valid       bool
}

type TFIDF struct {
	documents   []Document
	wc          map[string]float32
	idf         map[string]float32
	wordToIndex map[string]int
	indexToWord []string
}

func NewTransactionTFIDF() *TFIDF {
	return &TFIDF{
		documents: []Document{},
		wc:        map[string]float32{},
	}
}

func (p *TFIDF) indexWords() (mapping map[string]int, vectorSize int) {
	// Only calculate the index once. This way we can use it elsewhere if we need
	// to get information back out of the transform after we are done.
	if len(p.wordToIndex) == 0 {
		// Count the number of words that have more than one appearance, words that
		// only have a single appearance are not important to include in the vector.
		wordCount := 0
		for _, count := range p.wc {
			if count == 1 {
				continue
			}
			wordCount++
		}

		// Define the length of the vector and adjust it to be divisible by 32. This
		// will enable us to leverage SIMD in the future. By using 32 we are
		// compatible with both AVX and AVX512.
		vectorLength := wordCount + (calc.VectorWidthAlignment - (wordCount % calc.VectorWidthAlignment))
		p.wordToIndex = make(map[string]int)
		p.indexToWord = make([]string, vectorLength)
		allWords := make([]string, 0, len(p.wc))
		for word, count := range p.wc {
			if count == 1 {
				continue
			}
			allWords = append(allWords, word)
		}
		// Make the order of word indicies consistent between runs of the same data.
		sort.Strings(allWords)
		for index, word := range allWords {
			p.wordToIndex[word] = index
			p.indexToWord[index] = word
		}
	}

	return p.wordToIndex, len(p.indexToWord)
}

// Deprecated: Use Tokenize instead. This function is outdated.
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
	tokens := Tokenize(txn)
	// There may be more than len(tokens) words beacuse a single token may be
	// multiple words depending on the replacement table. However the word count
	// map will simply grow if this is the case.
	wordCounts := make(map[string]float32, len(tokens))
	for _, token := range tokens {
		for _, word := range token.Final {
			wordCounts[word]++
			p.wc[word]++
		}
	}

	tf := make(map[string]float32, len(wordCounts))
	for word, count := range wordCounts {
		tf[word] = count / float32(len(wordCounts))
	}

	p.documents = append(p.documents, Document{
		ID:          txn.TransactionId,
		Tokens:      tokens,
		Transaction: txn,
		TF:          tf,
		TFIDF:       map[string]float32{},
	})
}

func (p *TFIDF) GetDocuments(ctx context.Context) []Document {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	resultDocuments := make([]Document, 0, len(p.documents))
	docCount := float32(len(p.documents))
	p.idf = make(map[string]float32, len(p.wc))
	for word, count := range p.wc {
		p.idf[word] = float32(math.Log(float64(docCount / (count + 1))))
	}
	// Get a map of all the meaningful words and their index to use in the vector
	minified, vectorSize := p.indexWords()
	crumbs.Debug(span.Context(), "Organizing documents for DBSCAN clustering (transaction similarity)", map[string]any{
		"count":      len(p.documents),
		"vectorSize": vectorSize,
	})
	for i := range p.documents {
		// Get the current document we are working with
		document := p.documents[i]
		// Calculate the TFIDF for that document
		for word, tfValue := range document.TF {
			document.TFIDF[word] = tfValue * p.idf[word]
		}
		// Then create a vector of the words in the document name to use for the
		// DBSCAN clustering
		document.Vector = make([]float32, vectorSize)
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

		// Then build the sparse form of that vector for the clustering to use.
		// The TFIDF map already knows every word this document uses, so the
		// indicies can be collected straight from it instead of scanning the
		// whole vector looking for the few slots that are not zero.
		indices := make([]int32, 0, words)
		for word := range document.TFIDF {
			index, exists := minified[word]
			if !exists {
				continue
			}
			indices = append(indices, int32(index))
		}
		// Map iteration order is random and the distance calculation sums these
		// in the order they are given, so they have to be sorted for the result
		// to stay consistent between runs against the same data.
		sort.Slice(indices, func(i, j int) bool {
			return indices[i] < indices[j]
		})

		// The values are read back out of the vector after it has been normalized
		// so that they are exactly the floats the dense vector is holding, rather
		// than a separately divided copy that could differ in the last bit.
		document.Norm2, document.Signature = 0, 0
		document.Indices = make([]int32, 0, len(indices))
		document.Values = make([]float32, 0, len(indices))
		for _, index := range indices {
			value := document.Vector[index]
			// A word can end up with a tfidf value of zero. It contributes nothing
			// to the distance and would only weaken the signature.
			if value == 0 {
				continue
			}
			document.Indices = append(document.Indices, index)
			document.Values = append(document.Values, value)
			document.Norm2 += value * value
			document.Signature |= 1 << (uint64(index) % 64)
		}

		p.documents[i] = document
		// Then store the document back in
		if document.Valid {
			resultDocuments = append(resultDocuments, document)
		}
	}

	return resultDocuments
}
