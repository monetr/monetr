package recurring

import (
	"sort"

	"github.com/monetr/monetr/server/models"
)

type SimilarTransactionDetection interface {
	AddTransaction(txn *models.Transaction)
	DetectSimilarTransactions() []SimilarTransactionGroup
}

type SimilarTransactionGroup struct {
	Name    string   `json:"name"`
	Matches []uint64 `json:"matches"`
}

type SimilarTransactions_TFIDF_DBSCAN struct {
	preprocessor *TFIDF
	dbscan       *DBSCAN
}

func NewSimilarTransactions_TFIDF_DBSCAN() SimilarTransactionDetection {
	return &SimilarTransactions_TFIDF_DBSCAN{
		preprocessor: &TFIDF{
			documents: make([]Document, 0, 500),
			wc:        make(map[string]float32, 128),
		},
	}
}

func (s *SimilarTransactions_TFIDF_DBSCAN) AddTransaction(txn *models.Transaction) {
	s.preprocessor.AddTransaction(txn)
}

func (s *SimilarTransactions_TFIDF_DBSCAN) DetectSimilarTransactions() []SimilarTransactionGroup {
	datums := s.preprocessor.GetDatums()
	s.dbscan = NewDBSCAN(datums, Epsilon, MinNeighbors)
	result := s.dbscan.Calculate()
	similar := make([]SimilarTransactionGroup, len(result))

	for i, cluster := range result {
		group := SimilarTransactionGroup{
			Name:    "",
			Matches: make([]uint64, 0, len(cluster.Items)),
		}

		for index := range cluster.Items {
			datum, ok := s.dbscan.GetDatumByIndex(index)
			if !ok {
				// I don't know what kind of information would be helpful to include here since we cannot find the data
				// associated with the index anyway. But this would indicate a significant bug.
				panic("could not find a datum with an index in a resulting cluster")
			}

			// Add the transaction ID to the matches.
			group.Matches = append(group.Matches, datum.ID)
		}

		sort.Slice(group.Matches, func(i, j int) bool {
			return group.Matches[i] < group.Matches[j]
		})

		similar[i] = group
	}

	return similar
}
