package recurring

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecurringDetection(t *testing.T) {
	t.Run("amazon sample data", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2023, 12, 1, 9, 0, 0, 0, time.UTC))
		data := GetFixtures(t, "amazon_sample_data_1.json")

		result, err := DetectRecurringTransactions(context.Background(), clock, data)
		assert.NoError(t, err)

		j, err := json.MarshalIndent(result, "", "    ")
		require.NoError(t, err, "must be able to marshall result")

		fmt.Println(string(j))
	})

	t.Run("freshbooks sample data", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2022, 3, 1, 9, 0, 0, 0, time.UTC))
		data := GetFixtures(t, "monetr_freshbooks_data_1.json")

		result, err := DetectRecurringTransactions(context.Background(), clock, data)
		assert.NoError(t, err)

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
		detector := NewSimilarTransactions_TFIDF_DBSCAN(log)

		for i := range data {
			detector.AddTransaction(&data[i])
		}

		groups := detector.DetectSimilarTransactions(context.Background())
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

			recurringResult, err := DetectRecurringTransactions(context.Background(), clock, transactions)
			assert.NoError(t, err)

			if recurringResult.Best == nil || recurringResult.Best.StartDate.IsZero() {
				log.Infof("cluster: \"%s\" does not recur", group.Name)
				continue
			}

			log.Infof("cluster: \"%s\" does recur roughly every %d days", group.Name, recurringResult.Best.Frequency)

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
	})
}
