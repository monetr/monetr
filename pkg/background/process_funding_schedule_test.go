package background

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestProcessFundingScheduleJob_Run(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAPlaidLink(t, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)

		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, &bankAccount, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", false)
		fundingSchedule.NextOccurrence = util.MidnightInLocal(fundingSchedule.NextOccurrence.Add(-24 * time.Hour), timezone)
		testutils.MustDBUpdate(t, fundingSchedule)


		spendingRule := testutils.Must(t, models.NewRule, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR")
		spendingRule.DTStart(time.Now().Add(7 * 24 * time.Hour))
		nextDue := spendingRule.After(time.Now(), false)

		contributions := fundingSchedule.GetNumberOfContributionsBetween(time.Now(), nextDue)
		assert.NotZero(t, contributions, "must have at least one contribution, if this fails then this test is written wrong")

		spending := models.Spending{
			AccountId:              user.AccountId,
			Account:                user.Account,
			BankAccountId:          bankAccount.BankAccountId,
			BankAccount:            &bankAccount,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			FundingSchedule:        fundingSchedule,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Amazon",
			Description:            "Amazon Prime Subscription",
			TargetAmount:           1395,
			CurrentAmount:          697,
			UsedAmount:             0,
			RecurrenceRule:         spendingRule,
			LastRecurrence:         nil,
			NextRecurrence:         nextDue,
			NextContributionAmount: 100,
			IsBehind:               false,
			IsPaused:               false,
			DateCreated:            time.Now(),
		}
		testutils.MustDBInsert(t, &spending)

		handler := NewProcessFundingScheduleHandler(log, db)
		args := ProcessFundingScheduleArguments{
			AccountId:     fundingSchedule.AccountId,
			BankAccountId: bankAccount.BankAccountId,
			FundingScheduleIds: []uint64{
				fundingSchedule.FundingScheduleId,
			},
		}

		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "should run job successfully")
		testutils.MustHaveLogMessage(t, hook, "preparing to update 1 spending(s)")

		updatedSpending := testutils.MustDBRead(t, spending)
		assert.EqualValues(t, spending.CurrentAmount + spending.NextContributionAmount, updatedSpending.CurrentAmount, "current amount should have been incremented")
		assert.Greater(t, updatedSpending.NextContributionAmount, int64(0), "next contribution must be greater than 0")
	})

	t.Run("will fail for a fake account", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)

		handler := NewProcessFundingScheduleHandler(log, db)
		args := ProcessFundingScheduleArguments{
			AccountId:     math.MaxUint64,
			BankAccountId: 123,
			FundingScheduleIds: []uint64{
				123,
			},
		}

		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.EqualError(t, err, "failed to retrieve account: pg: no rows in result set")
		testutils.MustHaveLogMessage(t, hook, "could not retrieve account for funding schedule processing")
	})

	t.Run("will not process a future funding schedule", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		link := fixtures.GivenIHaveAPlaidLink(t, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, &bankAccount, "FREQ=DAILY;INTERVAL=1", false)

		handler := NewProcessFundingScheduleHandler(log, db)
		args := ProcessFundingScheduleArguments{
			AccountId:     fundingSchedule.AccountId,
			BankAccountId: bankAccount.BankAccountId,
			FundingScheduleIds: []uint64{
				fundingSchedule.FundingScheduleId,
			},
		}

		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "should run job successfully")
		testutils.MustHaveLogMessage(t, hook, "skipping processing funding schedule, it does not occur yet")
	})
}
