package recurring

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/monetr/monetr/server/models"
)

type Detection struct {
	timezone           *time.Location
	preprocessor       *TFIDF
	dbscan             *DBSCAN
	latestObservedDate time.Time
}

func NewRecurringTransactionDetection(timezone *time.Location) *Detection {
	return &Detection{
		timezone:           timezone,
		preprocessor:       NewTransactionTFIDF(),
		dbscan:             nil,
		latestObservedDate: time.Time{},
	}
}

func (d *Detection) AddTransaction(txn *models.Transaction) {
	d.preprocessor.AddTransaction(txn)
	if txn.Date.After(d.latestObservedDate) {
		d.latestObservedDate = txn.Date
	}
}

func (d *Detection) GetRecurringTransactions() []models.TransactionRecurring {
	type Hit struct {
		Window models.WindowType
		Time   time.Time
	}
	type Miss struct {
		Window models.WindowType
		Time   time.Time
	}
	type Transaction struct {
		ID       models.ID[models.Transaction]
		Name     string
		Merchant string
		Date     time.Time
		Amount   int64
	}

	d.dbscan = NewDBSCAN(d.preprocessor.GetDocuments(), Epsilon, MinNeighbors)
	result := d.dbscan.Calculate()
	bestScores := make([]models.TransactionRecurring, 0, len(result))

	for _, cluster := range result {
		clusterAmounts := map[int64]AmountCluster{}
		transactions := make([]*models.Transaction, 0, len(cluster.Items))
		for index := range cluster.Items {
			transaction := d.dbscan.dataset[index]
			transactions = append(transactions, transaction.Transaction)
			a, ok := clusterAmounts[transaction.Transaction.Amount]
			if !ok {
				a.IDs = make([]models.ID[models.Transaction], 0, 1)
				a.Amount = transaction.Transaction.Amount
			}
			a.IDs = append(a.IDs, transaction.ID)
			clusterAmounts[transaction.Transaction.Amount] = a
		}
		sort.Slice(transactions, func(i, j int) bool {
			return transactions[i].Date.Before(transactions[j].Date)
		})

		start, end := transactions[0].Date, transactions[len(transactions)-1].Date
		windows := GetWindowsForDate(transactions[0].Date, d.timezone)
		scores := make([]models.TransactionRecurring, 0, len(windows))
		for _, window := range windows {
			var lastAmount int64 = 0
			misses := make([]Miss, 0)
			hits := make([]Hit, 0, len(transactions))
			ids := make([]models.ID[models.Transaction], 0, len(transactions))
			amounts := make(map[int64]int, len(transactions))
			occurrences := window.Rule.Between(start.AddDate(0, 0, -window.Fuzzy), end.AddDate(0, 0, window.Fuzzy), false)
			if len(occurrences) == 1 {
				continue
			}
			for x := range occurrences {
				occurrence := occurrences[x]
				foundAny := false
				for i := range transactions {
					transaction := transactions[i]
					delta := math.Abs(transaction.Date.Sub(occurrence).Hours())
					fuzz := float64(window.Fuzzy) * 24
					if fuzz >= delta {
						foundAny = true
						hits = append(hits, Hit{
							Window: window.Type,
							Time:   occurrence,
						})
						amounts[transaction.Amount] += 1
						ids = append(ids, transaction.TransactionId)
						lastAmount = transaction.Amount
						continue
					}
				}
				if !foundAny {
					misses = append(misses, Miss{
						Window: window.Type,
						Time:   occurrence,
					})
				}
			}

			if len(hits) == 0 {
				continue
			}
			next := window.Rule.After(hits[len(hits)-1].Time, false)
			countHits := float32(len(hits))
			countMisses := float32(len(misses)) * 1.1
			countTxns := float32(len(transactions))
			ended := next.Before(d.latestObservedDate.AddDate(0, 0, -window.Fuzzy*2))
			latestTxn := transactions[len(transactions)-1]
			name := latestTxn.OriginalName
			if latestTxn.OriginalMerchantName != "" {
				name = latestTxn.OriginalMerchantName
			}

			scores = append(scores, models.TransactionRecurring{
				TransactionRecurringId: uuid.NewString(),
				Name:                   name,
				Window:                 window.Type,
				RuleSet:                &models.RuleSet{Set: *window.Rule},
				First:                  hits[0].Time,
				Last:                   hits[len(hits)-1].Time,
				Next:                   next,
				Ended:                  ended,
				Confidence:             (countHits - countMisses) / countTxns,
				Members:                ids,
				Amounts:                amounts,
				LastAmount:             lastAmount,
			})
		}

		if len(scores) > 0 {
			sort.Slice(scores, func(i, j int) bool {
				return scores[i].Confidence > scores[j].Confidence
			})

			if scores[0].Confidence > 0.65 {
				bestScores = append(bestScores, scores[0])
			}
		}
	}

	return bestScores
}

type AmountCluster struct {
	Amount int64
	IDs    []models.ID[models.Transaction]
}

func findBuckets(clusterAmounts map[int64]AmountCluster) []AmountCluster {
	amountsSorted := make([]AmountCluster, 0, len(clusterAmounts))
	for i := range clusterAmounts {
		amountsSorted = append(amountsSorted, clusterAmounts[i])
	}
	sort.Slice(amountsSorted, func(i, j int) bool {
		return amountsSorted[i].Amount < amountsSorted[j].Amount
	})

	if len(amountsSorted) <= 16 {
		return amountsSorted
	}

	bottom, top := amountsSorted[:len(amountsSorted)/2], amountsSorted[len(amountsSorted)/2:]
	bottomScore, topScore := 0, 0
	for _, bottomItem := range bottom {
		bottomScore += len(bottomItem.IDs)
	}
	for _, topItem := range top {
		topScore += len(topItem.IDs)
	}

	if bottomScore > topScore {
		return bottom
	}

	return top
}
