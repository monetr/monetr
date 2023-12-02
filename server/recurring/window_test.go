package recurring

import (
	"fmt"
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
			TransactionId: 291,
			Amount:        1500,
			OriginalCategories: []string{
				"Service",
				"Financial",
				"Accounting and Bookkeeping",
			},
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
		//data := GetFixtures(t, "Result_3.json")
		// data := GetFixtures(t, "full sample.json")
		var processor = &PreProcessor{
			documents: []Document{},
			wc:        map[string]int{},
			idf:       map[string]float64{},
		}
		for i := range data {
			processor.AddTransaction(&data[i])
		}

		processor.PostPrepareCalculations()

		assert.NotEmpty(t, processor.idf)

		dbscan := NewDBSCAN(processor.GetDatums(), 0.98, 1)
		result := dbscan.Calculate()
		assert.NotEmpty(t, result)

		for _, cluster := range result {
			transactions := make([]*models.Transaction, 0, len(cluster.Items))
			for index := range cluster.Items {
				transactions = append(transactions, dbscan.dataset[index].Transaction)
			}
			sort.Slice(transactions, func(i, j int) bool {
				return transactions[i].Date.Before(transactions[j].Date)
			})

			windows := GetWindowsForDate(transactions[0].Date, timezone)

			fmt.Println("base transaction:", transactions[0].OriginalName, "date:", transactions[0].Date)
			for _, transaction := range transactions[1:] {
				fmt.Println("\ttransaction:", transaction.OriginalName, "date:", transaction.Date)
				for _, window := range windows {
					days, ok := window.GetDeviation(transaction.Date)
					fmt.Println("\t\twindow:", window.Type, "aligns:", days, ok)
				}
			}
		}

	})
}
