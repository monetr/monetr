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
	Matches []uint64 `json:"matches"`
}

type SimilarTransactions_TFIDF_DBSCAN struct {
	tfidf  *TFIDF
	dbscan *DBSCAN
}

func NewSimilarTransactions_TFIDF_DBSCAN() SimilarTransactionDetection {
	return &SimilarTransactions_TFIDF_DBSCAN{
		tfidf: NewTransactionTFIDF(),
	}
}

func (s *SimilarTransactions_TFIDF_DBSCAN) AddTransaction(txn *models.Transaction) {
	s.tfidf.AddTransaction(txn)
}

func (s *SimilarTransactions_TFIDF_DBSCAN) DetectSimilarTransactions() []SimilarTransactionGroup {
	datums := s.tfidf.GetDocuments()
	s.dbscan = NewDBSCAN(datums, Epsilon, MinNeighbors)
	result := s.dbscan.Calculate()
	similar := make([]SimilarTransactionGroup, len(result))

	for i, cluster := range result {
		group := SimilarTransactionGroup{
			Matches: make([]uint64, 0, len(cluster.Items)),
		}

		// TODO I want to determine what the best name for a given cluster is, and naturally that name is somewhere in the
		// names of the transactions in that cluster. I have access to the TFIDF that generated this cluster at this point
		// in the code via `s.tfidf.idf` and the document for that item.
		// I think an approach might be to calculate another TFIDF given only the words in a single cluster. Then compare
		// the weights of those words against the weights of the same words from the parent TFIDF. This way we could
		// determine a relative weight against the whole. Words that are more uniquely identifying in the parent (Amazon for
		// example) will be less uniquely identifying in the sub cluster. Words that are the most unique in the sub cluster
		// we can probably assume to be useless, as if they are unique here then they are likely a reference number or
		// something that would always be unique.
		// If possible it would also be good to take into account the order of the terms for the final name. But I'm not
		// sure how important that will be yet.
		for index := range cluster.Items {
			datum, ok := s.dbscan.GetDocumentByIndex(index)
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
