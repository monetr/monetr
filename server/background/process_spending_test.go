package background

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestProcessSpendingJob_Run(t *testing.T) {
	t.Run("fix stale spending", func(t *testing.T) {
		clock := clock.NewMock()
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		timezone := testutils.MustEz(t, user.Account.GetTimezone)

		fundingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", clock.Now())
		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday",
			Description:            "Payday",
			RuleSet:                fundingRule,
			NextOccurrence:         fundingRule.After(clock.Now(), false),
			NextOccurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour)) // Allow past times.
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Test Stale Expense",
			Description:       "Description or something",
			TargetAmount:      5000,
			CurrentAmount:     5000,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true), // Make it so it recurs next in the past. (STALE)
			DateCreated:       clock.Now(),
		})

		handler := NewProcessSpendingHandler(log, db, clock)

		args := ProcessSpendingArguments{
			AccountId:     spending.AccountId,
			BankAccountId: spending.BankAccountId,
		}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "should run job successfully")
		testutils.MustHaveLogMessage(t, hook, "updating stale spending objects")

		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.Greater(t, updatedSpending.NextRecurrence, spending.NextRecurrence, "make sure the next recurrence field was updated")
	})
}
