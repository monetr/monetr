package background

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFundingScheduleJob_Run(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		timezone := testutils.MustEz(t, user.Account.GetTimezone)

		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", false)
		for fundingSchedule.NextOccurrence.After(clock.Now()) {
			fundingSchedule.NextOccurrence = fundingSchedule.NextOccurrence.AddDate(0, 0, -7)
		}
		testutils.MustDBUpdate(t, fundingSchedule)
		assert.Greater(t, clock.Now(), fundingSchedule.NextOccurrence, "next occurrence must be in the past")

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=2;BYDAY=FR", clock.Now())
		// spendingRule.DTStart(time.Now().Add(14 * 24 * time.Hour))
		nextDue := spendingRule.After(clock.Now(), false)

		timezone, err := user.Account.GetTimezone()
		require.NoError(t, err, "must get account timezone")

		contributions := fundingSchedule.GetNumberOfContributionsBetween(clock.Now(), nextDue, timezone)
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
			RuleSet:                spendingRule,
			LastRecurrence:         nil,
			NextRecurrence:         nextDue,
			NextContributionAmount: 100,
			IsBehind:               false,
			IsPaused:               false,
			DateCreated:            clock.Now(),
		}
		testutils.MustDBInsert(t, &spending)

		handler := NewProcessFundingScheduleHandler(log, db, clock)
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
		assert.EqualValues(t, spending.CurrentAmount+spending.NextContributionAmount, updatedSpending.CurrentAmount, "current amount should have been incremented")
		assert.Greater(t, updatedSpending.NextContributionAmount, int64(0), "next contribution must be greater than 0")
	})

	t.Run("will fail for a fake account", func(t *testing.T) {
		clock := clock.NewMock()
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)

		handler := NewProcessFundingScheduleHandler(log, db, clock)
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
		clock := clock.NewMock()
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=DAILY;INTERVAL=1", false)
		fundingSchedule.NextOccurrence = clock.Now().Add(1 * time.Hour).In(timezone)
		testutils.MustDBUpdate(t, fundingSchedule)

		handler := NewProcessFundingScheduleHandler(log, db, clock)
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
