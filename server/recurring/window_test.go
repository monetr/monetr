package recurring

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"testing"
	"time"

	"github.com/adrg/strutil/metrics"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestWindowGetDeviation(t *testing.T) {
	t.Run("weekly", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		date := time.Date(2023, 11, 19, 0, 0, 0, 0, timezone)
		weekly := windowWeekly(date)

		{ // Test the start date
			delta, ok := weekly.GetDeviation(date)
			assert.EqualValues(t, 0, delta, "input date should have a delta of 0")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test the next date
			next := date.AddDate(0, 0, 7)
			delta, ok := weekly.GetDeviation(next)
			assert.EqualValues(t, 0, delta, "next date should have a delta of 0")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test outside window
			next := date.AddDate(0, 0, 4)
			delta, ok := weekly.GetDeviation(next)
			assert.EqualValues(t, -1, delta)
			assert.False(t, ok)
		}

		{ // Test edge of window
			next := date.AddDate(0, 0, 2)
			delta, ok := weekly.GetDeviation(next)
			assert.EqualValues(t, 2, delta)
			assert.True(t, ok)
		}
	})

	t.Run("monthly", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		date := time.Date(2023, 11, 15, 0, 0, 0, 0, timezone)
		monthly := windowMonthly(date)

		{ // Test the start date
			delta, ok := monthly.GetDeviation(date)
			assert.EqualValues(t, 0, delta, "input date should have a delta of 0")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test the next date
			next := date.AddDate(0, 1, 0)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, 0, delta, "next date should have a delta of 0")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test the next date with one day after
			next := date.AddDate(0, 1, 1)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, 1, delta, "one day after the next should have a delta of 1")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test the next date with one day before
			next := date.AddDate(0, 1, -1)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, 1, delta, "one day before the next should have a delta of 1")
			assert.True(t, ok, "ok should be true when a date matches")
		}

		{ // Test before the start day
			next := date.AddDate(0, -1, 0)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, -1, delta, "invalid date should have a delta of -1")
			assert.False(t, ok, "ok should be false if the provided date comes before the start")
		}

		{ // Test outside the window
			next := date.AddDate(0, 0, 13)
			delta, ok := monthly.GetDeviation(next)
			assert.EqualValues(t, -1, delta, "should have a delta of -1 for an invalid day")
			assert.False(t, ok, "ok should be false if the provided date is outside the window")
		}
	})

	t.Run("with real data", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		data := GetFixtures(t, "monetr_sample_data_1.json")
		comparison := &transactionComparatorBase{
			impl: &metrics.JaroWinkler{
				CaseSensitive: false,
			},
		}
		searcher := &TransactionSearch{
			nameComparator:     comparison,
			merchantComparator: comparison,
		}
		baseline := models.Transaction{
			TransactionId:        "txn_291",
			Amount:               1500,
			Date:                 time.Date(2021, 7, 10, 0, 0, 0, 0, timezone),
			OriginalName:         "FreshBooks. Merchant name: Freshbooks",
			OriginalMerchantName: "FreshBooks",
		}

		windows := GetWindowsForDate(baseline.Date, timezone)

		result := searcher.FindSimilarTransactions(baseline, data)
		assert.NotEmpty(t, result, "should have found at least some transactions")

		type Match struct {
			Transaction models.Transaction
			Window      Window
			Delta       int
		}
		matches := make([]Match, 0, len(result))

		for _, txn := range result {
			for _, window := range windows {
				delta, ok := window.GetDeviation(txn.Date)
				if ok {
					matches = append(matches, Match{
						Transaction: txn,
						Window:      window,
						Delta:       delta,
					})
				}
			}
		}

		assert.NotEmpty(t, matches)
	})
}

func TestWindowExperiment(t *testing.T) {
	t.Run("with cluster", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		data := GetFixtures(t, "monetr_sample_data_1.json")
		// data := GetFixtures(t, "Result_3.json")
		// data := GetFixtures(t, "full sample.json")
		processor := NewTransactionTFIDF()
		latest := time.Time{}
		for i := range data {
			processor.AddTransaction(&data[i])
			if data[i].Date.After(latest) {
				latest = data[i].Date
			}
		}

		dbscan := NewDBSCAN(processor.GetDocuments(context.Background()), 0.98, 1)
		result := dbscan.Calculate(context.Background())
		assert.NotEmpty(t, result)

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
		type Score struct {
			Group              string
			Window             models.WindowType
			Start              time.Time
			Last               time.Time
			Next               time.Time
			Hits               int
			Misses             int
			Transactions       int
			Confidence         float64
			Ended              bool
			TransactionMatches []Transaction
		}

		bestScores := make([]Score, 0, len(result))

		for _, cluster := range result {
			transactions := make([]*models.Transaction, 0, len(cluster.Items))
			for index := range cluster.Items {
				transactions = append(transactions, dbscan.dataset[index].Transaction)
			}
			sort.Slice(transactions, func(i, j int) bool {
				return transactions[i].Date.Before(transactions[j].Date)
			})

			start, end := transactions[0].Date, transactions[len(transactions)-1].Date
			windows := GetWindowsForDate(transactions[0].Date, timezone)
			scores := make([]Score, 0, len(windows))
			for _, window := range windows {
				misses := make([]Miss, 0)
				hits := make([]Hit, 0, len(transactions))
				ids := make([]Transaction, 0, len(transactions))
				occurrences := window.Rule.Between(start.AddDate(0, 0, -window.Fuzzy), end.AddDate(0, 0, window.Fuzzy), false)
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
							ids = append(ids, Transaction{
								ID:       transaction.TransactionId,
								Name:     transaction.OriginalName,
								Merchant: transaction.OriginalMerchantName,
								Date:     transaction.Date,
								Amount:   transaction.Amount,
							})
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
				scores = append(scores, Score{
					Group:              transactions[0].OriginalName,
					Window:             window.Type,
					Hits:               len(hits),
					Misses:             len(misses),
					Transactions:       len(transactions),
					Start:              hits[0].Time,
					Last:               hits[len(hits)-1].Time,
					Next:               next,
					Ended:              next.Before(latest.AddDate(0, 0, window.Fuzzy)),
					TransactionMatches: ids,
				})
			}

			sort.Slice(scores, func(i, j int) bool {
				hitsA := float64(scores[i].Hits)
				missesA := float64(scores[i].Misses) * 1.1
				txnsA := float64(scores[i].Transactions)

				hitsB := float64(scores[j].Hits)
				missesB := float64(scores[j].Misses) * 1.1
				txnsB := float64(scores[j].Transactions)

				accuracyA := (hitsA - missesA) / txnsA
				accuracyB := (hitsB - missesB) / txnsB
				scores[i].Confidence = accuracyA
				scores[j].Confidence = accuracyB
				return accuracyA > accuracyB
			})

			if scores[0].Confidence > 0.65 && !scores[0].Ended {
				bestScores = append(bestScores, scores[0])
			}
		}

		j, err := json.MarshalIndent(bestScores, "", "    ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(j))
	})
}
