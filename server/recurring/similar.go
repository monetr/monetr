package recurring

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/models"
	"github.com/sirupsen/logrus"
)

const NumberOfMostValuableWords = 2

type SimilarTransactionDetection interface {
	AddTransaction(txn *models.Transaction)
	DetectSimilarTransactions(ctx context.Context) []models.TransactionCluster
}

type SimilarTransactions_TFIDF_DBSCAN struct {
	log    *logrus.Entry
	tfidf  *TFIDF
	dbscan *DBSCAN
}

func NewSimilarTransactions_TFIDF_DBSCAN(log *logrus.Entry) SimilarTransactionDetection {
	return &SimilarTransactions_TFIDF_DBSCAN{
		log:   log,
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
	similar := make([]models.TransactionCluster, 0, len(result))

	for _, cluster := range result {
		group := models.TransactionCluster{
			Members: make([]models.ID[models.Transaction], len(cluster.Items)),
		}
		group.TransactionClusterId = models.NewID(&group)

		indicies := s.tfidf.indexToWord
		mostValuableIndicies := make([]struct {
			Word         string
			OriginalWord string
			Placement    int
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

			sort.Slice(datum.Tokens, func(i, j int) bool {
				return rankWordComposition(datum.Tokens[i].Original) < rankWordComposition(datum.Tokens[j].Original)
			})

			for wordIndex, wordValue := range datum.Vector {
				tracker := mostValuableIndicies[wordIndex]
				tracker.Word = indicies[wordIndex]
				tracker.Value += wordValue

				if tracker.OriginalWord == "" {
					// TODO, There are multiple tokens for the same word in this array and
					// we may not necessarily select the "best" one. We probably want to
					// eventually prioritize tokens that do not have as many capital
					// characters as they are more likely to be the regular name. For
					// example Jetbrains versus JETBRAINS in a transaction name. We also
					// want to calculate the placement based on the AVERAGE of the
					// indexes, as the first token we see probably is not accurate and if
					// the transaction name changes AT ALL over time then the order will
					// get fucked up.
					for _, originalToken := range datum.Tokens {
						for tokenIndex, tokenWord := range originalToken.Final {
							if strings.EqualFold(indicies[wordIndex], tokenWord) {
								if len(originalToken.Final) > 1 {
									tracker.OriginalWord = originalToken.Equivalent[tokenIndex]
								} else {
									tracker.OriginalWord = originalToken.Original
								}
								tracker.Placement = originalToken.Index
								break
							}
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

		// Post processing steps for the similar transaction cluster...

		{ // Sort the members of the cluster by their transaction date.
			sort.SliceStable(items, func(i, j int) bool {
				return items[i].Date.After(items[j].Date)
			})
			for i := range items {
				group.Members[i] = items[i].ID
			}
		}

		{ // Calculate a consistent ID and a "name" for the cluster
			sort.SliceStable(mostValuableIndicies, func(i, j int) bool {
				return mostValuableIndicies[i].Value > mostValuableIndicies[j].Value
			})

			group.Debug = make([]models.TransactionClusterDebugItem, 0, len(mostValuableIndicies))
			for _, item := range mostValuableIndicies {
				if item.Value == 0 {
					break
				}

				group.Debug = append(group.Debug, models.TransactionClusterDebugItem{
					Word:      item.OriginalWord,
					Sanitized: item.Word,
					Order:     item.Placement,
					Value:     item.Value,
				})
			}

			calculateRankings(&group)
			calculatedMerchantName(&group)
			calculateSignature(&group)
		}

		// Don't return a transaction cluster with no name, this can happen somehow
		// but I'm still debugging exactly how.
		if group.Name == "" {
			s.log.
				WithFields(logrus.Fields{
					"bug":     true,
					"members": group.Members,
				}).
				Warn("transaction cluster was calculated to not have a name, investigate!")
			continue
		}

		similar = append(similar, group)
	}

	return similar
}

// calculateRankings takes all of the values of the most valuable tokens in a
// transaction cluster and creates a ranking value that is more normalized in
// order to better select values.
func calculateRankings(group *models.TransactionCluster) {
	vectorSize := len(group.Debug) + (16 - (len(group.Debug) % 16))
	rankings := make([]float32, vectorSize)
	for i := range group.Debug {
		rankings[i] = group.Debug[i].Value * group.Debug[i].Value
	}
	calc.NormalizeVector32(rankings)
	for i := range group.Debug {
		group.Debug[i].Rank = rankings[i]
	}
}

func calculatedMerchantName(group *models.TransactionCluster) {
	maximum, minimum := group.Debug[0].Rank, group.Debug[len(group.Debug)-1].Rank
	var cutoff float32 = 0.8 // I want values in the top 90% of rankings
	threshold := minimum + (maximum-minimum)*cutoff
	items := make([]models.TransactionClusterDebugItem, 0, len(group.Debug))
	for i := range group.Debug {
		item := group.Debug[i]
		if item.Rank > threshold {
			items = append(items, item)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Order < items[j].Order
	})

	group.Merchant = items
}

func calculateSignature(group *models.TransactionCluster) {
	consistentId := sha256.New()
	slug := make([]string, len(group.Merchant))
	hashSlug := make([]string, len(group.Merchant))
	for i, part := range group.Merchant {
		slug[i] = part.Word
		hashSlug[i] = strings.ToLower(part.Sanitized)
	}
	// This way the hash is consistent if the values of the top values changes
	// slightly. As long as the words themselves are the same the hash stays the
	// same.
	sort.Strings(hashSlug)
	consistentId.Write([]byte(strings.Join(hashSlug, "-")))
	group.Name = strings.TrimSpace(strings.Join(slug, " "))
	group.Signature = hex.EncodeToString(consistentId.Sum(nil))
}

// rankWordComposition takes a word and returns a score based on the composition
// of the word. Words that are all uppercase are the lowest score and rank 0.
// Words that are all lower case are rank 1, and words that are title case rank
// 2. This is used to select the best words from the array of tokens for a
// transaction cluster to produce the most ideal merchant name.
func rankWordComposition(word string) int {
	switch {
	case word == strings.ToUpper(word):
		return 0
	case word == strings.ToLower(word):
		return 1
	default:
		// Title case
		if len(word) > 1 &&
			unicode.IsUpper(rune(word[0])) &&
			word[1:] == strings.ToLower(word[1:]) {
			return 2
		}
		// TODO This might need to be more nuanced in the future. Like what about
		// words like GitHub or YouTube?
		// Mixed case is the same as lower
		return 1
	}
}
