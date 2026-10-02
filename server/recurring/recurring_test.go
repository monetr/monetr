package recurring

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/similar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecurringDetection(t *testing.T) {
	t.Run("amazon sample data", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2023, 12, 1, 9, 0, 0, 0, time.UTC))
		data := GetFixtures(t, "amazon_sample_data_1.json")

		results, err := DetectRecurringTransactions(t.Context(), clock, time.UTC, data)
		assert.NoError(t, err)

		j, err := json.MarshalIndent(results, "", "    ")
		require.NoError(t, err, "must be able to marshall result")

		fmt.Println(string(j))
	})

	t.Run("freshbooks sample data", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2022, 3, 1, 9, 0, 0, 0, time.UTC))
		data := GetFixtures(t, "monetr_freshbooks_data_1.json")

		results, err := DetectRecurringTransactions(t.Context(), clock, time.UTC, data)
		assert.NoError(t, err)
		require.Len(t, results, 1, "freshbooks charges are all debits")
		result := results[0]

		j, err := json.MarshalIndent(result, "", "    ")
		require.NoError(t, err, "must be able to marshall result")

		fmt.Println(string(j))

		assert.EqualValues(t, 30, result.Best.Frequency, "should recurr every 30 days")
	})

	t.Run("larger sample data", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2023, 12, 1, 9, 0, 0, 0, time.UTC))
		// First build out several transaction clusters
		data := GetFixtures(t, "monetr_sample_data_1.json")
		log := testutils.GetLog(t)
		detector := similar.NewSimilarTransactions_TFIDF_DBSCAN(log)

		for i := range data {
			detector.AddTransaction(&data[i])
		}

		groups := detector.DetectSimilarTransactions(t.Context())
		assert.NotEmpty(t, groups, "must return an array of groups of similar transactions")
		for _, group := range groups {
			if len(group.Members) < 3 {
				continue
			}

			assert.NotEmpty(t, group.Members, "a groups matches should not be empty!")
			assert.NotEmpty(t, group.Name, "a groups name should not be empty!")
			assert.NotEmpty(t, group.Signature, "a groups signature should not be empty!")

			transactions := make([]models.Transaction, 0, len(group.Members))
		MemberLoop:
			for _, memberId := range group.Members {
				for i := range data {
					transaction := data[i]
					if transaction.TransactionId == memberId {
						transactions = append(transactions, transaction)
						continue MemberLoop
					}
				}
			}

			recurringResults, err := DetectRecurringTransactions(t.Context(), clock, time.UTC, transactions)
			assert.NoError(t, err)

			for _, recurringResult := range recurringResults {
				if recurringResult.Best == nil || recurringResult.Best.StartDate.IsZero() {
					log.Info(fmt.Sprintf("cluster: %q does not recur", group.Name))
					continue
				}

				log.Info(fmt.Sprintf("cluster: %q does recur roughly every %d days", group.Name, recurringResult.Best.Frequency))

				switch strings.ToLower(group.Name) {
				case "freshbooks":
					// Should recur monthly
					assert.Contains(t, []int{30, 31}, recurringResult.Best.Frequency)
				case "github inc":
					// Should recur monthly
					assert.Contains(t, []int{30, 31}, recurringResult.Best.Frequency)
				case "sentry":
					// Should recur monthly
					assert.Contains(t, []int{30, 31}, recurringResult.Best.Frequency)
				case "treasury courant elliot":
					// Should recur twice a month
					assert.Contains(t, []int{15, 16}, recurringResult.Best.Frequency)
				}
			}
		}
	})

	t.Run("yearly renewal", func(t *testing.T) {
		// Three renewals on the same day every year, like a yearly subscription.
		transactions := []models.Transaction{
			{TransactionId: "txn_0", Amount: 5985, Date: time.Date(2024, 4, 23, 0, 0, 0, 0, time.UTC)},
			{TransactionId: "txn_1", Amount: 5985, Date: time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC)},
			{TransactionId: "txn_2", Amount: 7188, Date: time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC)},
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), time.UTC, transactions)
		require.NoError(t, err, "must be able to check the transactions for recurrence")
		require.Len(t, results, 1, "all of the renewals are debits")
		result := results[0]
		require.NotNil(t, result.Best, "a yearly renewal must be detected as recurring")
		assert.EqualValues(t, 365, result.Best.Frequency, "should recur every year")
		assert.Len(t, result.Members, 3, "every renewal should be part of the result")
	})

	t.Run("visits roughly a year apart are not yearly", func(t *testing.T) {
		// Three visits to the same place. The longer gap is 446 days, which a loose
		// tolerance would accept as about a year, but the gaps are 446 and 273 days
		// and that is nothing like a yearly charge.
		transactions := []models.Transaction{
			{TransactionId: "txn_0", Amount: 1520, Date: time.Date(2023, 3, 10, 0, 0, 0, 0, time.UTC)},
			{TransactionId: "txn_1", Amount: 829, Date: time.Date(2024, 5, 29, 0, 0, 0, 0, time.UTC)},
			{TransactionId: "txn_2", Amount: 1537, Date: time.Date(2025, 2, 26, 0, 0, 0, 0, time.UTC)},
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), time.UTC, transactions)
		require.NoError(t, err, "must be able to check the transactions for recurrence")
		require.Len(t, results, 1, "all of the visits are debits")
		result := results[0]
		assert.Nil(t, result.Best, "irregular visits must not be detected as recurring")
	})
}
