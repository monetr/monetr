package recurring_jobs

import (
	"fmt"
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
		diff := DiffTransactionRecurring(t.Context(), nil, nil, nil, now, time.UTC, "bac_test", "tcl_test")
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
					{
						TransactionId: "txn_0",
						Amount:        800,
						Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_1",
						Amount:        800,
						Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_2",
						Amount:        1000,
						Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, nil, now, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the new recurring transaction")
		assert.Empty(t, diff.DeleteRecurringIds, "should not delete anything")

		item := diff.UpsertRecurring[0]
		assert.False(t, item.TransactionRecurringId.IsZero(), "a new recurring transaction gets its ID from the diff so members can point at it")
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
					{
						TransactionId: "txn_0",
						Amount:        800,
						Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_1",
						Amount:        800,
						Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_2",
						Amount:        800,
						Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), existing, results, nil, now, time.UTC, "bac_test", "tcl_test")
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

		diff := DiffTransactionRecurring(t.Context(), existing, results, nil, now, time.UTC, "bac_test", "tcl_test")
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
					{
						TransactionId: "txn_0",
						Amount:        800,
						Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_1",
						Amount:        800,
						Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_2",
						Amount:        800,
						Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), existing, results, nil, now, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the debits")
		assert.NotEqualValues(t, "txrc_credit", diff.UpsertRecurring[0].TransactionRecurringId, "the debits are new and must not reuse the credit ID")
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
					{
						TransactionId: "txn_0",
						Amount:        800,
						Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_1",
						Amount:        800,
						Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_2",
						Amount:        800,
						Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, nil, now, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should still upsert a recurring transaction that ended")
		assert.True(t, diff.UpsertRecurring[0].Ended, "should be ended")
	})

	t.Run("rule is evaluated in the account's timezone", func(t *testing.T) {
		// Tokyo is ahead of UTC, so midnight on the 8th in Tokyo is still the 7th
		// in UTC. If the rule were evaluated in UTC then every occurrence would be
		// a day late and the last transaction would look like it was missed.
		tokyo, err := time.LoadLocation("Asia/Tokyo")
		require.NoError(t, err, "must be able to load the timezone")

		now := time.Date(2026, 7, 1, 9, 0, 0, 0, tokyo)
		ruleset := "DTSTART:20251231T150000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=8"
		members := make([]models.Transaction, 6)
		for i := range members {
			members[i] = models.Transaction{
				TransactionId: models.ID[models.Transaction](fmt.Sprintf("txn_%d", i)),
				Amount:        800,
				Date:          time.Date(2026, time.Month(i+1), 8, 0, 0, 0, 0, tokyo),
			}
		}
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.DebitDirection,
				Best: &recurring.Frequency{
					Frequency:  30,
					Confidence: 0.9,
				},
				RuleSet: testutils.Must(t, models.NewRuleSet, ruleset),
				Members: members,
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, nil, now, tokyo, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the recurring transaction")

		item := diff.UpsertRecurring[0]
		assert.False(t, item.Ended, "should not have ended, the next charge is a week away")
		assert.Equal(t, time.Date(2026, 7, 8, 0, 0, 0, 0, tokyo), item.Next.In(tokyo), "next should be the 8th in Tokyo")
		assert.Equal(t, ruleset, item.RuleSet.String(), "the stored ruleset should stay in UTC")
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
					{
						TransactionId: "txn_0",
						Amount:        -500000,
						Date:          time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_1",
						Amount:        -500000,
						Date:          time.Date(2026, 1, 30, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_2",
						Amount:        -500000,
						Date:          time.Date(2026, 2, 13, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_3",
						Amount:        -500000,
						Date:          time.Date(2026, 2, 27, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, nil, now, time.UTC, "bac_test", "tcl_test")
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
					{
						TransactionId: "txn_0",
						Amount:        800,
						Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_1",
						Amount:        800,
						Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_1",
						Amount:        800,
						Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_2",
						Amount:        800,
						Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, nil, now, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the recurring transaction")
		assert.Equal(t, map[int64]int{800: 3}, diff.UpsertRecurring[0].Amounts, "the duplicate member should only be counted once")
	})

	t.Run("members point at the new recurring transaction", func(t *testing.T) {
		now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		transactions := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        800,
				Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_1",
				Amount:        800,
				Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_2",
				Amount:        800,
				Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			},
		}
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
				Members: transactions,
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, transactions, now, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the new recurring transaction")
		recurringId := diff.UpsertRecurring[0].TransactionRecurringId
		require.False(t, recurringId.IsZero(), "the new recurring transaction must have an ID")

		require.Len(t, diff.UpdateMembers, 3, "every member should be updated")
		for _, member := range diff.UpdateMembers {
			require.NotNil(t, member.TransactionRecurringId, "member must point at a recurring transaction")
			assert.Equal(t, recurringId, *member.TransactionRecurringId, "member must point at the new recurring transaction")
		}
	})

	t.Run("only members that changed are updated", func(t *testing.T) {
		now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		existing := []models.TransactionRecurring{
			{
				TransactionRecurringId: "txrc_debit",
				AccountId:              "acct_test",
				BankAccountId:          "bac_test",
				TransactionClusterId:   "tcl_test",
				Direction:              models.DebitDirection,
			},
		}
		// txn_0 already points at the recurring transaction, txn_1 and txn_2 are new
		// members, and txn_3 used to be a member but is an outlier now.
		transactions := []models.Transaction{
			{
				TransactionId:          "txn_0",
				Amount:                 800,
				Date:                   time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
				TransactionRecurringId: new(models.ID[models.TransactionRecurring]("txrc_debit")),
			},
			{
				TransactionId: "txn_1",
				Amount:        800,
				Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_2",
				Amount:        800,
				Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId:          "txn_3",
				Amount:                 800,
				Date:                   time.Date(2026, 3, 29, 0, 0, 0, 0, time.UTC),
				TransactionRecurringId: new(models.ID[models.TransactionRecurring]("txrc_debit")),
			},
		}
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
				Members: transactions[:3],
			},
		}

		diff := DiffTransactionRecurring(t.Context(), existing, results, transactions, now, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the existing recurring transaction")

		updated := make(map[models.ID[models.Transaction]]*models.ID[models.TransactionRecurring], len(diff.UpdateMembers))
		for _, member := range diff.UpdateMembers {
			updated[member.TransactionId] = member.TransactionRecurringId
		}
		assert.Len(t, updated, 3, "only the members that changed should be updated")
		assert.NotContains(t, updated, models.ID[models.Transaction]("txn_0"), "txn_0 already points at the recurring transaction")
		if assert.Contains(t, updated, models.ID[models.Transaction]("txn_1")) {
			assert.EqualValues(t, "txrc_debit", *updated["txn_1"], "txn_1 should point at the recurring transaction")
		}
		if assert.Contains(t, updated, models.ID[models.Transaction]("txn_2")) {
			assert.EqualValues(t, "txrc_debit", *updated["txn_2"], "txn_2 should point at the recurring transaction")
		}
		if assert.Contains(t, updated, models.ID[models.Transaction]("txn_3")) {
			assert.Nil(t, updated["txn_3"], "txn_3 is not a member anymore so it should not point at anything")
		}
	})

	t.Run("members are cleared when the direction no longer recurs", func(t *testing.T) {
		now := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		existing := []models.TransactionRecurring{
			{
				TransactionRecurringId: "txrc_debit",
				AccountId:              "acct_test",
				BankAccountId:          "bac_test",
				TransactionClusterId:   "tcl_test",
				Direction:              models.DebitDirection,
			},
		}
		transactions := []models.Transaction{
			{
				TransactionId:          "txn_0",
				Amount:                 800,
				Date:                   time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
				TransactionRecurringId: new(models.ID[models.TransactionRecurring]("txrc_debit")),
			},
			{
				TransactionId:          "txn_1",
				Amount:                 800,
				Date:                   time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
				TransactionRecurringId: new(models.ID[models.TransactionRecurring]("txrc_debit")),
			},
		}
		results := []recurring.RecurringTransactionResult{
			{
				Direction: models.DebitDirection,
				Best:      nil,
			},
		}

		diff := DiffTransactionRecurring(t.Context(), existing, results, transactions, now, time.UTC, "bac_test", "tcl_test")
		assert.Equal(t, []models.ID[models.TransactionRecurring]{"txrc_debit"}, diff.DeleteRecurringIds, "should delete the recurring transaction")
		require.Len(t, diff.UpdateMembers, 2, "both old members should be updated")
		for _, member := range diff.UpdateMembers {
			assert.Nil(t, member.TransactionRecurringId, "old members should not point at anything")
		}
	})

	t.Run("both directions recur", func(t *testing.T) {
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
		// txn_2 used to be counted with the credits by mistake, now it is a debit
		// member and should move over to the new debit recurring transaction.
		debits := []models.Transaction{
			{
				TransactionId: "txn_0",
				Amount:        800,
				Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId: "txn_1",
				Amount:        800,
				Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
			},
			{
				TransactionId:          "txn_2",
				Amount:                 800,
				Date:                   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
				TransactionRecurringId: new(models.ID[models.TransactionRecurring]("txrc_credit")),
			},
		}
		credits := []models.Transaction{
			{
				TransactionId:          "txn_3",
				Amount:                 -800,
				Date:                   time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC),
				TransactionRecurringId: new(models.ID[models.TransactionRecurring]("txrc_credit")),
			},
			{
				TransactionId:          "txn_4",
				Amount:                 -800,
				Date:                   time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC),
				TransactionRecurringId: new(models.ID[models.TransactionRecurring]("txrc_credit")),
			},
			{
				TransactionId:          "txn_5",
				Amount:                 -800,
				Date:                   time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC),
				TransactionRecurringId: new(models.ID[models.TransactionRecurring]("txrc_credit")),
			},
		}
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
				Members: debits,
			},
			{
				Direction: models.CreditDirection,
				Best: &recurring.Frequency{
					Frequency:  30,
					Confidence: 0.85,
				},
				RuleSet: testutils.Must(
					t,
					models.NewRuleSet,
					"DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=25",
				),
				Members: credits,
			},
		}
		transactions := append(append([]models.Transaction{}, debits...), credits...)

		diff := DiffTransactionRecurring(t.Context(), existing, results, transactions, now, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 2, "should upsert both directions")
		assert.Empty(t, diff.DeleteRecurringIds, "should not delete anything")

		byDirection := make(map[models.Direction]models.TransactionRecurring, len(diff.UpsertRecurring))
		for _, item := range diff.UpsertRecurring {
			byDirection[item.Direction] = item
		}
		require.Contains(t, byDirection, models.DebitDirection, "should have a debit recurring transaction")
		require.Contains(t, byDirection, models.CreditDirection, "should have a credit recurring transaction")
		debitId := byDirection[models.DebitDirection].TransactionRecurringId
		assert.EqualValues(t, "txrc_credit", byDirection[models.CreditDirection].TransactionRecurringId, "the credits should keep the existing ID")
		assert.False(t, debitId.IsZero(), "the debits should get a new ID")
		assert.NotEqualValues(t, "txrc_credit", debitId, "the debits must not reuse the credit ID")

		updated := make(map[models.ID[models.Transaction]]*models.ID[models.TransactionRecurring], len(diff.UpdateMembers))
		for _, member := range diff.UpdateMembers {
			updated[member.TransactionId] = member.TransactionRecurringId
		}
		assert.Len(t, updated, 3, "only the debit members should be updated, the credits already point at the right one")
		for _, id := range []models.ID[models.Transaction]{"txn_0", "txn_1", "txn_2"} {
			if assert.Contains(t, updated, id) && assert.NotNil(t, updated[id]) {
				assert.Equal(t, debitId, *updated[id], "debit member should point at the debit recurring transaction")
			}
		}
	})

	t.Run("not ended within the slack after a missed occurrence", func(t *testing.T) {
		// The last charge was March 20th, so April 20th was expected. Monthly gets
		// 15 days of slack, so it isn't ended until after May 5th.
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
					{
						TransactionId: "txn_0",
						Amount:        800,
						Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_1",
						Amount:        800,
						Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_2",
						Amount:        800,
						Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		}

		withinSlack := time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC)
		diff := DiffTransactionRecurring(t.Context(), nil, results, nil, withinSlack, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the recurring transaction")
		assert.False(t, diff.UpsertRecurring[0].Ended, "should not be ended while still within the slack")

		pastSlack := time.Date(2026, 5, 5, 0, 0, 1, 0, time.UTC)
		diff = DiffTransactionRecurring(t.Context(), nil, results, nil, pastSlack, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the recurring transaction")
		assert.True(t, diff.UpsertRecurring[0].Ended, "should be ended once past the slack")
	})

	t.Run("first and last come from the dates not the order", func(t *testing.T) {
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
					{
						TransactionId: "txn_1",
						Amount:        800,
						Date:          time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_2",
						Amount:        1000,
						Date:          time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
					},
					{
						TransactionId: "txn_0",
						Amount:        700,
						Date:          time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC),
					},
				},
			},
		}

		diff := DiffTransactionRecurring(t.Context(), nil, results, nil, now, time.UTC, "bac_test", "tcl_test")
		require.Len(t, diff.UpsertRecurring, 1, "should upsert the recurring transaction")
		item := diff.UpsertRecurring[0]
		assert.Equal(t, time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), item.First, "first should be the earliest member")
		assert.Equal(t, time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), item.Last, "last should be the latest member")
		assert.EqualValues(t, 1000, item.LastAmount, "last amount should be from the latest member")
	})

	t.Run("deleted ids are sorted", func(t *testing.T) {
		now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		existing := []models.TransactionRecurring{
			{
				TransactionRecurringId: "txrc_b",
				AccountId:              "acct_test",
				BankAccountId:          "bac_test",
				TransactionClusterId:   "tcl_test",
				Direction:              models.DebitDirection,
			},
			{
				TransactionRecurringId: "txrc_a",
				AccountId:              "acct_test",
				BankAccountId:          "bac_test",
				TransactionClusterId:   "tcl_test",
				Direction:              models.CreditDirection,
			},
		}

		// Map iteration order is random, so run it a few times to catch an
		// unsorted result.
		for range 10 {
			diff := DiffTransactionRecurring(t.Context(), existing, nil, nil, now, time.UTC, "bac_test", "tcl_test")
			assert.Empty(t, diff.UpsertRecurring, "should not upsert anything")
			assert.Equal(t, []models.ID[models.TransactionRecurring]{"txrc_a", "txrc_b"}, diff.DeleteRecurringIds, "should delete both in order")
		}
	})
}

func TestWindowType(t *testing.T) {
	cases := []struct {
		name      string
		frequency int
		rule      string
		expected  models.WindowType
	}{
		{
			name:      "weekly",
			frequency: 7,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=WEEKLY;INTERVAL=1;BYDAY=FR",
			expected:  models.WeeklyWindowType,
		},
		{
			name:      "every two weeks",
			frequency: 14,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=WEEKLY;INTERVAL=2;BYDAY=FR",
			expected:  models.BiWeeklyWindowType,
		},
		{
			name:      "1st and the 15th",
			frequency: 15,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1,15",
			expected:  models.FirstAndFifteenthWindowType,
		},
		{
			name:      "15th and the last day",
			frequency: 15,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
			expected:  models.FifteenthAndLastWindowType,
		},
		{
			name:      "monthly",
			frequency: 30,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15",
			expected:  models.MonthlyWindowType,
		},
		{
			name:      "every two months",
			frequency: 60,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=2;BYMONTHDAY=15",
			expected:  models.BiMonthlyWindowType,
		},
		{
			name:      "quarterly",
			frequency: 90,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=15",
			expected:  models.QuarterlyWindowType,
		},
		{
			name:      "yearly",
			frequency: 365,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=YEARLY;INTERVAL=1;BYMONTH=4;BYMONTHDAY=23",
			expected:  models.YearlyWindowType,
		},
		{
			name:      "unknown frequency falls back to monthly",
			frequency: 45,
			rule:      "DTSTART:20260101T000000Z\nRRULE:FREQ=DAILY;INTERVAL=45",
			expected:  models.MonthlyWindowType,
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			result := recurring.RecurringTransactionResult{
				Best: &recurring.Frequency{
					Frequency: item.frequency,
				},
				RuleSet: testutils.Must(t, models.NewRuleSet, item.rule),
			}
			assert.Equal(t, item.expected, windowType(result), "window type should match the frequency")
		})
	}
}
