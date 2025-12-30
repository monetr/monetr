package recurring

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/sirupsen/logrus"
)

func PPrint(thing any) {
	j, err := json.MarshalIndent(thing, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(j))
}

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
	ID models.ID[models.Transaction]

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
	indicies := s.tfidf.indexToWord
	for _, cluster := range result {
		group := models.TransactionCluster{
			Members: make([]models.ID[models.Transaction], len(cluster.Items)),
		}
		group.TransactionClusterId = models.NewID(&group)

		mostValuableIndicies := make([]struct {
			Word         string
			OriginalWord string
			Placement    float32
			Value        float32
			Count        int
		}, len(indicies))

		items := make([]memberItem, 0, len(cluster.Items))
		centroid := make([]float32, len(datums[0].Vector))

		for index := range cluster.Items {
			datum, ok := s.dbscan.GetDocumentByIndex(index)
			if !ok {
				// I don't know what kind of information would be helpful to include
				// here since we cannot find the data associated with the index anyway.
				// But this would indicate a significant bug.
				panic("could not find a datum with an index in a resulting cluster")
			}

			// Build the centroid coordinates
			for i := range datum.Vector {
				centroid[i] += datum.Vector[i]
			}

			// Add the transaction ID to the matches.
			items = append(items, memberItem{
				ID:   datum.ID,
				Date: datum.Transaction.Date,
			})

			// This is done for every item. Every word in every item in a cluster is
			// compiled into a single set of tracking metrics.
			for wordIndex, wordValue := range datum.Vector {
				// Skip over words that arent apart of this item.
				if wordValue == 0 {
					continue
				}

				// Grab the existing value for this word's index in the tracker.
				tracker := mostValuableIndicies[wordIndex]
				// Make sure that the word is stored
				tracker.Word = indicies[wordIndex]
				// And that we increment that words value.
				tracker.Value += wordValue
				tracker.Count++
				// Then consider all of the transactions that contained the same word,
				// what was the position of that word within those transactions. We
				// want the average position so that way its relative to other words.
				var position, count float32
				for _, originalToken := range datum.Tokens {
					for _, tokenWord := range originalToken.Final {
						if strings.EqualFold(indicies[wordIndex], tokenWord) {
							position += float32(originalToken.Index)
							count++
							break
						}
					}
				}
				tracker.Placement = position / count
				mostValuableIndicies[wordIndex] = tracker
			}
		}

		size := float32(len(cluster.Items))
		for i := range centroid {
			centroid[i] /= size
		}

		minimumDistance := float32(math.MaxFloat32)
		centerIndex := -1
		// If we have multiple documents that are considered "center" use the
		// document with the lowest ID.
		minId := "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"
		for index := range cluster.Items {
			datum, ok := s.dbscan.GetDocumentByIndex(index)
			if !ok {
				panic("could not find a datum with an index in a resulting cluster")
			}

			distance := calc.EuclideanDistance32(datum.Vector, centroid)
			if distance < minimumDistance {
				minimumDistance = distance
				centerIndex = index
				minId = string(datum.ID)
			} else if distance == minimumDistance && string(datum.ID) < minId {
				minimumDistance = distance
				centerIndex = index
				minId = string(datum.ID)
			}
		}

		{ // Based on the center most datum, calculate the merchant
			datum, _ := s.dbscan.GetDocumentByIndex(centerIndex)

			sort.Slice(datum.Tokens, func(i, j int) bool {
				return rankWordComposition(datum.Tokens[i].Original) < rankWordComposition(datum.Tokens[j].Original)
			})

			// Then we take the center most transaction in the cluster and we use the
			// original words from that transaction in our most valuable indicies
			// tracker. This way the most valuable words remain consistent as long as
			// the center transaction does not change significantly over time.
			for wordIndex := range datum.Vector {
				tracker := mostValuableIndicies[wordIndex]
				for _, originalToken := range datum.Tokens {
					for tokenIndex, tokenWord := range originalToken.Final {
						if strings.EqualFold(indicies[wordIndex], tokenWord) {
							if len(originalToken.Final) > 1 {
								tracker.OriginalWord = originalToken.Equivalent[tokenIndex]
							} else {
								tracker.OriginalWord = originalToken.Original
							}
							break
						}
					}
				}

				mostValuableIndicies[wordIndex] = tracker
			}
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
				// TODO: This is technically not perfect, I'm still digging into why
				// some values can end up being negative but it can happen. And when it
				// does this sort ends up fucking things up where the highest value
				// items will have a value of 0 because the real items have a negative
				// value. So sorting by the absolute value should be good enough for
				// now but long term I need to improve this.
				// See https://github.com/monetr/monetr/issues/2833 for more info.
				return myownsanity.AbsFloat32(mostValuableIndicies[i].Value) >
					myownsanity.AbsFloat32(mostValuableIndicies[j].Value)
			})

			group.Debug = make([]models.TransactionClusterDebugItem, 0, len(mostValuableIndicies))
			for _, item := range mostValuableIndicies {
				if item.Value == 0 {
					continue
				}

				group.Debug = append(group.Debug, models.TransactionClusterDebugItem{
					Word:      item.OriginalWord,
					Sanitized: item.Word,
					Order:     item.Placement,
					Value:     item.Value,
					Count:     float32(item.Count),
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
		rankings[i] = (group.Debug[i].Value / group.Debug[i].Count) * (group.Debug[i].Value / group.Debug[i].Count)
	}
	calc.NormalizeVector32(rankings)
	for i := range group.Debug {
		group.Debug[i].Rank = rankings[i]
	}
}

func calculatedMerchantName(group *models.TransactionCluster) {
	if len(group.Debug) == 0 {
		j, _ := json.MarshalIndent(group, "", "  ")
		panic(fmt.Sprintf("Transaction cluster does not contain valuable words?\n%s", string(j)))
	}
	maximum, minimum := group.Debug[0].Rank, group.Debug[len(group.Debug)-1].Rank
	var cutoff float32 = 0.75 // I want values in the top 75% of rankings
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
