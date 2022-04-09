package background

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestProcessSpendingJob_Run(t *testing.T) {
	t.Run("fix stale spending", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAPlaidLink(t, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		fundingRule := testutils.Must(t, models.NewRule, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1")
		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:      bankAccount.AccountId,
			BankAccountId:  bankAccount.BankAccountId,
			Name:           "Payday",
			Description:    "Payday",
			Rule:           fundingRule,
			NextOccurrence: fundingRule.After(time.Now(), false),
		})

		spendingRule := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO")
		spendingRule.DTStart(time.Now().Add(-8 * 24 * time.Hour)) // Allow past times.
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Test Stale Expense",
			Description:       "Description or something",
			TargetAmount:      5000,
			CurrentAmount:     5000,
			RecurrenceRule:    spendingRule,
			NextRecurrence:    spendingRule.Before(time.Now(), true), // Make it so it recurs next in the past. (STALE)
			DateCreated:       time.Now(),
		})

		handler := NewProcessSpendingHandler(log, db)

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
