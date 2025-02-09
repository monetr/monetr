package recurring

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
)

type SimilarTransactionDetection interface {
	AddTransaction(txn *models.Transaction)
	DetectSimilarTransactions(ctx context.Context) []models.TransactionCluster
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

type memberItem struct {
	ID   models.ID[models.Transaction]
	Date time.Time
}

func (s *SimilarTransactions_TFIDF_DBSCAN) DetectSimilarTransactions(
	ctx context.Context,
) []models.TransactionCluster {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	datums := s.tfidf.GetDocuments(span.Context())
	s.dbscan = NewDBSCAN(datums, Epsilon, MinNeighbors)
	result := s.dbscan.Calculate(span.Context())
	similar := make([]models.TransactionCluster, len(result))

	for i, cluster := range result {
		group := models.TransactionCluster{
			Members: make([]models.ID[models.Transaction], len(cluster.Items)),
		}
		group.TransactionClusterId = models.NewID(&group)

		// TODO I want to determine what the best name for a given cluster is, and
		// naturally that name is somewhere in the names of the transactions in that
		// cluster. I have access to the TFIDF that generated this cluster at this
		// point in the code via `s.tfidf.idf` and the document for that item. I
		// think an approach might be to calculate another TFIDF given only the
		// words in a single cluster. Then compare the weights of those words
		// against the weights of the same words from the parent TFIDF. This way we
		// could determine a relative weight against the whole. Words that are more
		// uniquely identifying in the parent (Amazon for example) will be less
		// uniquely identifying in the sub cluster. Words that are the most unique
		// in the sub cluster we can probably assume to be useless, as if they are
		// unique here then they are likely a reference number or something that
		// would always be unique. If possible it would also be good to take into
		// account the order of the terms for the final name. But I'm not sure how
		// important that will be yet.
		// =========================================================================
		// In the mean time I'm going to use the most common merchant name or the
		// name of the transaction with the highest ID.

		indicies := s.tfidf.indexToWord
		mostValuableIndicies := make([]struct {
			Word         string
			OriginalWord string
			Value        float32
		}, len(indicies))

		items := make([]memberItem, 0, len(cluster.Items))
		for index := range cluster.Items {
			datum, ok := s.dbscan.GetDocumentByIndex(index)
			if !ok {
				// I don't know what kind of information would be helpful to include
				// here since we cannot find the data associated with the index anyway.
				// But this would indicate a significant bug.
				panic("could not find a datum with an index in a resulting cluster")
			}

			for wordIndex, wordValue := range datum.Vector {
				tracker := mostValuableIndicies[wordIndex]
				tracker.Word = indicies[wordIndex]
				tracker.Value += wordValue

				if tracker.OriginalWord == "" {
					for _, originalWord := range datum.UpperParts {
						if strings.EqualFold(indicies[wordIndex], originalWord) {
							tracker.OriginalWord = originalWord
							break
						}
					}
				}

				mostValuableIndicies[wordIndex] = tracker
			}

			// Add the transaction ID to the matches.
			items = append(items, memberItem{
				ID:   datum.ID,
				Date: datum.Transaction.Date,
			})
		}

		sort.SliceStable(mostValuableIndicies, func(i, j int) bool {
			return mostValuableIndicies[i].Value > mostValuableIndicies[j].Value
		})

		consistentId := sha512.New()
		slug := make([]string, 0, 2)
		for _, valuableWord := range mostValuableIndicies[0:cap(slug)] {
			if valuableWord.Value == 0 {
				continue
			}
			slug = append(slug, valuableWord.OriginalWord)
			consistentId.Write([]byte(valuableWord.Word))
		}

		sort.SliceStable(items, func(i, j int) bool {
			return items[i].Date.After(items[j].Date)
		})

		for i := range items {
			group.Members[i] = items[i].ID
		}

		group.Name = strings.Join(slug, " ")
		group.Signature = hex.EncodeToString(consistentId.Sum(nil))

		similar[i] = group
	}

	return similar
}
