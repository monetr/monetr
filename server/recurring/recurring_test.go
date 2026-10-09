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
			{
				TransactionId: "txn_0",
				Amount:        5985,
				Date:          time.Date(2024, 4, 23, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_1",
				Amount:        5985,
				Date:          time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_2",
				Amount:        7188,
				Date:          time.Date(2026, 4, 23, 0, 0, 0, 0, time.UTC),
			},
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
			{
				TransactionId: "txn_0",
				Amount:        1520,
				Date:          time.Date(2023, 3, 10, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_1",
				Amount:        829,
				Date:          time.Date(2024, 5, 29, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_2",
				Amount:        1537,
				Date:          time.Date(2025, 2, 26, 0, 0, 0, 0, time.UTC),
			},
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), time.UTC, transactions)
		require.NoError(t, err, "must be able to check the transactions for recurrence")
		require.Len(t, results, 1, "all of the visits are debits")
		result := results[0]
		assert.Nil(t, result.Best, "irregular visits must not be detected as recurring")
	})

	t.Run("debits and credits", func(t *testing.T) {
		central, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load the timezone")

		// A cluster that has both a monthly charge on the 8th and a monthly refund
		// on the 22nd, each direction should be its own recurring result.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        1858,
				Date:          time.Date(2026, 1, 8, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_1",
				Amount:        -500,
				Date:          time.Date(2026, 1, 22, 0, 0, 0, 0, central),
			},
			// 8th was a Sunday
			{
				TransactionId: "txn_2",
				Amount:        1858,
				Date:          time.Date(2026, 2, 9, 0, 0, 0, 0, central),
			},
			// 22nd was a Sunday
			{
				TransactionId: "txn_3",
				Amount:        -500,
				Date:          time.Date(2026, 2, 23, 0, 0, 0, 0, central),
			},
			// 8th was a Sunday
			{
				TransactionId: "txn_4",
				Amount:        1858,
				Date:          time.Date(2026, 3, 9, 0, 0, 0, 0, central),
			},
			// 22nd was a Sunday
			{
				TransactionId: "txn_5",
				Amount:        -500,
				Date:          time.Date(2026, 3, 23, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_6",
				Amount:        1858,
				Date:          time.Date(2026, 4, 8, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_7",
				Amount:        -500,
				Date:          time.Date(2026, 4, 22, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_8",
				Amount:        1858,
				Date:          time.Date(2026, 5, 8, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_9",
				Amount:        -500,
				Date:          time.Date(2026, 5, 22, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_10",
				Amount:        1858,
				Date:          time.Date(2026, 6, 8, 0, 0, 0, 0, central),
			},
			{
				TransactionId: "txn_11",
				Amount:        -500,
				Date:          time.Date(2026, 6, 22, 0, 0, 0, 0, central),
			},
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), central, transactions)
		require.NoError(t, err, "must be able to check the transactions for recurrence")
		require.Len(t, results, 2, "there should be a result for each direction")

		debit := results[0]
		assert.Equal(t, models.DebitDirection, debit.Direction, "debits should come first")
		require.NotNil(t, debit.Best, "the monthly charge must be detected as recurring")
		assert.EqualValues(t, 30, debit.Best.Frequency, "the charge should recur every month")
		require.NotNil(t, debit.RuleSet, "a recurring result must have a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8", debit.RuleSet.GetRRule().OrigOptions.RRuleString(), "the charge should be on the 8th")
		for _, member := range debit.Members {
			assert.Positive(t, member.Amount, "debit members must only be debits")
		}

		credit := results[1]
		assert.Equal(t, models.CreditDirection, credit.Direction, "credits should come second")
		require.NotNil(t, credit.Best, "the monthly refund must be detected as recurring")
		assert.EqualValues(t, 30, credit.Best.Frequency, "the refund should recur every month")
		require.NotNil(t, credit.RuleSet, "a recurring result must have a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=22", credit.RuleSet.GetRRule().OrigOptions.RRuleString(), "the refund should be on the 22nd")
		for _, member := range credit.Members {
			assert.Negative(t, member.Amount, "credit members must only be credits")
		}
	})

	t.Run("one direction skipped", func(t *testing.T) {
		// A monthly charge with a single refund mixed in, there aren't enough
		// credits to detect anything so there is only a result for the debits.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        1440,
				Date:          time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_1",
				Amount:        1440,
				Date:          time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_2",
				Amount:        -1440,
				Date:          time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_3",
				Amount:        1440,
				Date:          time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_4",
				Amount:        1440,
				Date:          time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
			},
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), time.UTC, transactions)
		require.NoError(t, err, "must be able to check the transactions for recurrence")
		require.Len(t, results, 1, "only the debits have enough transactions")
		assert.Equal(t, models.DebitDirection, results[0].Direction, "the result should be for the debits")
		require.NotNil(t, results[0].Best, "the monthly charge must be detected as recurring")
		assert.Len(t, results[0].Members, 4, "the refund must not be a member")
	})

	t.Run("not enough transactions", func(t *testing.T) {
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        2000,
				Date:          time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_1",
				Amount:        -2000,
				Date:          time.Date(2026, 3, 26, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_2",
				Amount:        2000,
				Date:          time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC),
			},
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), time.UTC, transactions)
		assert.NoError(t, err, "not having enough transactions should not be an error")
		assert.Empty(t, results, "there should be no results")
	})

	t.Run("zero amounts", func(t *testing.T) {
		// Transactions with no amount have no direction, so these three don't count
		// towards either one.
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        0,
				Date:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_1",
				Amount:        0,
				Date:          time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_2",
				Amount:        0,
				Date:          time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), time.UTC, transactions)
		assert.NoError(t, err, "zero amounts should not be an error")
		assert.Empty(t, results, "there should be no results")
	})

	t.Run("monthly from a rule", func(t *testing.T) {
		central, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load the timezone")

		// Generate two years of charges on the 10th of every month, then the rule
		// we detect from them should generate those same dates again.
		source := inTimezone(testutils.Must(
			t,
			models.NewRuleSet,
			"DTSTART:20240101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10",
		), central)
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, central)
		end := time.Date(2025, 12, 31, 0, 0, 0, 0, central)
		dates := source.Between(start, end, true)
		require.Len(t, dates, 24, "should have generated a charge for every month")

		transactions := make([]models.Transaction, len(dates))
		for i, date := range dates {
			transactions[i] = models.Transaction{
				TransactionId: models.ID[models.Transaction](fmt.Sprintf("txn_%d", i)),
				Amount:        1999,
				Date:          date,
			}
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), central, transactions)
		require.NoError(t, err, "must be able to check the transactions for recurrence")
		require.Len(t, results, 1, "all of the charges are debits")
		result := results[0]
		require.NotNil(t, result.Best, "charges generated from a monthly rule must be detected as recurring")
		assert.EqualValues(t, 30, result.Best.Frequency, "should recur every month")
		assert.Len(t, result.Members, len(dates), "every charge should be a member")
		require.NotNil(t, result.RuleSet, "a recurring result must have a ruleset")
		assert.Equal(t, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=10", result.RuleSet.GetRRule().OrigOptions.RRuleString(), "should detect the same rule the charges were generated from")
		// The rule is only built from the most recent charges, so it starts in 2025
		// instead of 2024
		assert.Equal(t, dates[len(dates)-ruleSetRecentTransactions:], inTimezone(result.RuleSet, central).Between(start, end, true), "the detected rule should generate the same dates")
	})

	t.Run("every other week from a rule", func(t *testing.T) {
		central, err := time.LoadLocation("America/Chicago")
		require.NoError(t, err, "must be able to load the timezone")

		// Generate a year of paychecks every other Friday, this crosses daylight
		// savings time in both directions.
		source := inTimezone(testutils.Must(
			t,
			models.NewRuleSet,
			"DTSTART:20250103T060000Z\nRRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=FR",
		), central)
		start := time.Date(2025, 1, 1, 0, 0, 0, 0, central)
		end := time.Date(2025, 12, 31, 0, 0, 0, 0, central)
		dates := source.Between(start, end, true)
		require.Len(t, dates, 26, "should have generated a paycheck every other week")

		transactions := make([]models.Transaction, len(dates))
		for i, date := range dates {
			transactions[i] = models.Transaction{
				TransactionId: models.ID[models.Transaction](fmt.Sprintf("txn_%d", i)),
				Amount:        -250000,
				Date:          date,
			}
		}

		results, err := DetectRecurringTransactions(t.Context(), clock.NewMock(), central, transactions)
		require.NoError(t, err, "must be able to check the transactions for recurrence")
		require.Len(t, results, 1, "all of the paychecks are credits")
		result := results[0]
		assert.Equal(t, models.CreditDirection, result.Direction, "paychecks should be credits")
		require.NotNil(t, result.Best, "paychecks generated from an every other week rule must be detected as recurring")
		assert.EqualValues(t, 14, result.Best.Frequency, "should recur every other week")
		require.NotNil(t, result.RuleSet, "a recurring result must have a ruleset")
		assert.Equal(t, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", result.RuleSet.GetRRule().OrigOptions.RRuleString(), "should detect the same rule the paychecks were generated from")

		// The detected rule starts on its earliest member, so only compare the
		// dates from there on.
		detected := inTimezone(result.RuleSet, central)
		expected := make([]time.Time, 0, len(dates))
		for _, date := range dates {
			if !date.Before(detected.GetDTStart()) {
				expected = append(expected, date)
			}
		}
		assert.Equal(t, expected, detected.Between(start, end, true), "the detected rule should generate the same dates")
	})
}
