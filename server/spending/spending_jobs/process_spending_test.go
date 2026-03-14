package spending_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/spending/spending_jobs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProcessSpending(t *testing.T) {
	t.Run("fix stale spending", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t)

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
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
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
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			// First time, no notifications should be enqueued and we should not have
			// an error.
			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should run job successfully")
		}

		testutils.MustHaveLogMessage(t, hook, "updating stale spending objects")

		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.Greater(t, updatedSpending.NextRecurrence, spending.NextRecurrence, "make sure the next recurrence field was updated")
	})
}
