package recurring_jobs

import (
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/recurring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffTransactionRecurring(t *testing.T) {
	t.Run("both empty", func(t *testing.T) {
		now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		diff := DiffTransactionRecurring(t.Context(), nil, nil, now, "bac_test", "tcl_test")
		assert.Empty(t, diff.UpsertRecurring)
		assert.Empty(t, diff.DeleteRecurringIds)
	})

	t.Run("no existing all new", func(t *testing.T) {
		now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.DebitDirection,
				Best: &recurring.Frequency{
					Frequency:  30,
					Confidence: 0.9,
				},
				RuleSet: testutils.Must(
					t,
					models.NewRuleSet,
					"DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=20",
				),
				Members: []models.Transaction{
					{TransactionId: "txn_0", Amount: 800, Date: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_1", Amount: 800, Date: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_2", Amount: 1000, Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, now, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the new recurring transaction")
		assert.Empty(t, diff.DeleteRecurringIds, "should not delete anything")

		item := diff.UpsertRecurring[0]
		assert.True(t, item.TransactionRecurringId.IsZero(), "a new recurring transaction gets its ID when it is inserted")
		assert.EqualValues(t, "bac_test", item.BankAccountId)
		assert.EqualValues(t, "tcl_test", item.TransactionClusterId)
		assert.Equal(t, models.DebitDirection, item.Direction)
		assert.Equal(t, models.MonthlyWindowType, item.Window)
		assert.Equal(t, time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), item.First, "first should be the earliest member")
		assert.Equal(t, time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), item.Last, "last should be the latest member")
		assert.Equal(t, time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC), item.Next, "next should be the next occurrence after now")
		assert.False(t, item.Ended, "the next occurrence hasn't happened yet so it hasn't ended")
		assert.Equal(t, float32(0.9), item.Confidence)
		assert.Equal(t, map[int64]int{800: 2, 1000: 1}, item.Amounts, "should count each amount")
		assert.EqualValues(t, 1000, item.LastAmount, "last amount should be from the latest member")
	})

	t.Run("existing direction keeps its id", func(t *testing.T) {
		now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		createdAt := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
		existing := []models.TransactionRecurring{
			{
				TransactionRecurringId: "txrc_debit",
				AccountId:              "acct_test",
				BankAccountId:          "bac_test",
				TransactionClusterId:   "tcl_test",
				Direction:              models.DebitDirection,
				CreatedAt:              createdAt,
			},
		}
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.DebitDirection,
				Best: &recurring.Frequency{
					Frequency:  30,
					Confidence: 0.8,
				},
				RuleSet: testutils.Must(
					t,
					models.NewRuleSet,
					"DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=20",
				),
				Members: []models.Transaction{
					{TransactionId: "txn_0", Amount: 800, Date: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_1", Amount: 800, Date: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_2", Amount: 800, Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), existing, results, now, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the existing recurring transaction")
		assert.Empty(t, diff.DeleteRecurringIds, "should not delete anything")
		assert.EqualValues(t, "txrc_debit", diff.UpsertRecurring[0].TransactionRecurringId, "should keep the existing ID")
		assert.Equal(t, createdAt, diff.UpsertRecurring[0].CreatedAt, "should keep the existing created at")
	})

	t.Run("direction that no longer recurs is deleted", func(t *testing.T) {
		now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		existing := []models.TransactionRecurring{
			{
				TransactionRecurringId: "txrc_debit",
				AccountId:              "acct_test",
				BankAccountId:          "bac_test",
				TransactionClusterId:   "tcl_test",
				Direction:              models.DebitDirection,
			},
		}
		// There were enough debits to check, but they don't recur anymore.
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.DebitDirection,
				Best:      nil,
			},
		}

		diff := DiffTransactionRecurring(t.Context(), existing, results, now, "bac_test", "tcl_test")
		assert.Empty(t, diff.UpsertRecurring, "should not upsert anything")
		assert.Equal(t, []models.ID[models.TransactionRecurring]{"txrc_debit"}, diff.DeleteRecurringIds, "should delete the recurring transaction")
	})

	t.Run("direction missing from the results is deleted", func(t *testing.T) {
		now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		existing := []models.TransactionRecurring{
			{
				TransactionRecurringId: "txrc_credit",
				AccountId:              "acct_test",
				BankAccountId:          "bac_test",
				TransactionClusterId:   "tcl_test",
				Direction:              models.CreditDirection,
			},
		}
		// Only the debits recur now, the credits didn't even have enough
		// transactions to be checked.
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.DebitDirection,
				Best: &recurring.Frequency{
					Frequency:  30,
					Confidence: 0.9,
				},
				RuleSet: testutils.Must(
					t,
					models.NewRuleSet,
					"DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=20",
				),
				Members: []models.Transaction{
					{TransactionId: "txn_0", Amount: 800, Date: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_1", Amount: 800, Date: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_2", Amount: 800, Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), existing, results, now, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the debits")
		assert.True(t, diff.UpsertRecurring[0].TransactionRecurringId.IsZero(), "the debits are new")
		assert.Equal(t, models.DebitDirection, diff.UpsertRecurring[0].Direction)
		assert.Equal(t, []models.ID[models.TransactionRecurring]{"txrc_credit"}, diff.DeleteRecurringIds, "should delete the credits")
	})

	t.Run("ended when past the next expected occurrence", func(t *testing.T) {
		// The last charge was in March, so April 20th was expected and it is now
		// well past that.
		now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.DebitDirection,
				Best: &recurring.Frequency{
					Frequency:  30,
					Confidence: 0.9,
				},
				RuleSet: testutils.Must(
					t,
					models.NewRuleSet,
					"DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=20",
				),
				Members: []models.Transaction{
					{TransactionId: "txn_0", Amount: 800, Date: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_1", Amount: 800, Date: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_2", Amount: 800, Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, now, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should still upsert a recurring transaction that ended")
		assert.True(t, diff.UpsertRecurring[0].Ended, "should be ended")
	})

	t.Run("twice a month window type", func(t *testing.T) {
		now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.CreditDirection,
				Best: &recurring.Frequency{
					Frequency:  15,
					Confidence: 0.86,
				},
				RuleSet: testutils.Must(
					t,
					models.NewRuleSet,
					"DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
				),
				Members: []models.Transaction{
					{TransactionId: "txn_0", Amount: -500000, Date: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_1", Amount: -500000, Date: time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_2", Amount: -500000, Date: time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_3", Amount: -500000, Date: time.Date(2026, 2, 27, 0, 0, 0, 0, time.UTC)},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, now, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the recurring transaction")
		assert.Equal(t, models.FifteenthAndLastWindowType, diff.UpsertRecurring[0].Window, "should be the 15th and the last day of the month")
	})

	t.Run("duplicate members are only counted once", func(t *testing.T) {
		now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.DebitDirection,
				Best: &recurring.Frequency{
					Frequency:  30,
					Confidence: 0.9,
				},
				RuleSet: testutils.Must(
					t,
					models.NewRuleSet,
					"DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=20",
				),
				Members: []models.Transaction{
					{TransactionId: "txn_0", Amount: 800, Date: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_1", Amount: 800, Date: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_1", Amount: 800, Date: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)},
					{TransactionId: "txn_2", Amount: 800, Date: time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, now, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the recurring transaction")
		assert.Equal(t, map[int64]int{800: 3}, diff.UpsertRecurring[0].Amounts, "the duplicate member should only be counted once")
	})
}
