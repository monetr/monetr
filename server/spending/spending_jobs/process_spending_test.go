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

func TestProcessSpendingCron(t *testing.T) {
	t.Run("no stale spending to process", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			context.EXPECT().Log().Return(log).AnyTimes()

			err := spending_jobs.ProcessSpendingCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("enqueues jobs for bank accounts with stale spending", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

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
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Test Stale Expense",
			Description:       "Should trigger enqueue",
			TargetAmount:      5000,
			CurrentAmount:     0,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(spending_jobs.ProcessSpending),
				gomock.Any(),
				gomock.Eq(spending_jobs.ProcessSpendingArguments{
					AccountId:     bankAccount.AccountId,
					BankAccountId: bankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			err := spending_jobs.ProcessSpendingCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("does not enqueue for paused spending", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

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

		// Create a paused spending object with a stale next recurrence. The
		// cron query filters on is_paused = false, so this should not trigger
		// an enqueue.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Paused Stale Expense",
			Description:       "Should not trigger enqueue",
			TargetAmount:      5000,
			CurrentAmount:     0,
			IsPaused:          true,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			context.EXPECT().Log().Return(log).AnyTimes()

			err := spending_jobs.ProcessSpendingCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})
}

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

	t.Run("no stale spending objects", func(t *testing.T) {
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

		// Create a spending object whose next recurrence is in the future. This
		// spending is not stale and should not be processed.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Future Expense",
			Description:       "Not stale",
			TargetAmount:      5000,
			CurrentAmount:     0,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.After(clock.Now(), false),
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should succeed with no stale spending")
		}

		testutils.MustHaveLogMessage(t, hook, "no stale spending object were updated")

		// Verify the spending was not modified.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.Equal(t, spending.NextRecurrence.UTC(), updatedSpending.NextRecurrence.UTC(),
			"next recurrence should not have changed")
	})

	t.Run("paused spending is skipped", func(t *testing.T) {
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

		// Create a stale spending that is paused. The job should skip it even
		// though its next recurrence is in the past.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Paused Expense",
			Description:       "This expense is paused",
			TargetAmount:      5000,
			CurrentAmount:     0,
			IsPaused:          true,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should succeed with paused spending")
		}

		testutils.MustHaveLogMessage(t, hook, "no stale spending object were updated")

		// Verify the spending was NOT updated.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.Equal(t, spending.NextRecurrence.UTC(), updatedSpending.NextRecurrence.UTC(),
			"next recurrence should not have changed for paused spending")
		assert.Nil(t, updatedSpending.LastRecurrence,
			"last recurrence should remain nil for paused spending")
	})

	t.Run("spending behind on funding", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		timezone := testutils.MustEz(t, user.Account.GetTimezone)

		// Fund monthly on the 15th and last day of the month.
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

		// Weekly spending with zero current amount. If the spending recurs before
		// the next funding event, the spending is behind because there are not
		// enough funds allocated to cover it.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Behind Expense",
			Description:       "Not enough funds allocated",
			TargetAmount:      5000,
			CurrentAmount:     0,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should process spending behind on funding")
		}

		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.Greater(t, updatedSpending.NextRecurrence, spending.NextRecurrence,
			"next recurrence should have been advanced")
		assert.True(t, updatedSpending.IsBehind,
			"spending should be marked as behind when current amount cannot cover upcoming events")
	})

	t.Run("spending not behind when adequately funded", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
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

		// Weekly spending that is stale but has enough current amount to cover
		// all spending events before the next funding event.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Funded Expense",
			Description:       "Has enough funds",
			TargetAmount:      5000,
			CurrentAmount:     50000,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should process adequately funded spending")
		}

		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.Greater(t, updatedSpending.NextRecurrence, spending.NextRecurrence,
			"next recurrence should have been advanced")
		assert.False(t, updatedSpending.IsBehind,
			"spending should not be behind when adequately funded")
	})

	t.Run("last recurrence is set when next recurrence advances", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
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
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		originalNextRecurrence := spendingRule.Before(clock.Now(), true)
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Recurrence Tracking Expense",
			Description:       "Verify last recurrence is set",
			TargetAmount:      5000,
			CurrentAmount:     5000,
			RuleSet:           spendingRule,
			NextRecurrence:    originalNextRecurrence,
			CreatedAt:         clock.Now(),
		})

		assert.Nil(t, spending.LastRecurrence,
			"last recurrence should initially be nil")

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should process spending successfully")
		}

		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.NotNil(t, updatedSpending.LastRecurrence,
			"last recurrence should be set after processing stale spending")
		assert.Equal(t, originalNextRecurrence.UTC(), updatedSpending.LastRecurrence.UTC(),
			"last recurrence should equal the previous next recurrence")
		assert.Greater(t, updatedSpending.NextRecurrence, *updatedSpending.LastRecurrence,
			"next recurrence should be after the last recurrence")
	})

	t.Run("next contribution amount is calculated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
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

		// Stale spending with zero current amount. After processing, the
		// contribution amount should be non-zero because funds are needed.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Contribution Expense",
			Description:            "Needs contribution calculated",
			TargetAmount:           5000,
			CurrentAmount:          0,
			NextContributionAmount: 0,
			RuleSet:                spendingRule,
			NextRecurrence:         spendingRule.Before(clock.Now(), true),
			CreatedAt:              clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should calculate contribution amount")
		}

		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.Greater(t, updatedSpending.NextContributionAmount, int64(0),
			"next contribution amount should be non-zero when spending needs funding")
	})

	t.Run("fully funded spending has zero contribution", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
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

		// Stale spending that has far more funds than needed. The contribution
		// amount should remain zero because no additional funding is required.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Over Funded Expense",
			Description:            "Already has plenty of funds",
			TargetAmount:           5000,
			CurrentAmount:          50000,
			NextContributionAmount: 1000,
			RuleSet:                spendingRule,
			NextRecurrence:         spendingRule.Before(clock.Now(), true),
			CreatedAt:              clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should process over-funded spending")
		}

		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(0), updatedSpending.NextContributionAmount,
			"next contribution amount should be zero when over-funded")
	})

	t.Run("multiple stale spending objects are updated", func(t *testing.T) {
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

		spendingRule1 := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule1.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending1 := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Stale Expense One",
			Description:       "First stale expense",
			TargetAmount:      5000,
			CurrentAmount:     5000,
			RuleSet:           spendingRule1,
			NextRecurrence:    spendingRule1.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		spendingRule2 := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", clock.Now())
		spendingRule2.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending2 := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Stale Expense Two",
			Description:       "Second stale expense",
			TargetAmount:      3000,
			CurrentAmount:     3000,
			RuleSet:           spendingRule2,
			NextRecurrence:    spendingRule2.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending1.AccountId,
					BankAccountId: spending1.BankAccountId,
				},
			)
			assert.NoError(t, err, "should process multiple stale spending objects")
		}

		testutils.MustHaveLogMessage(t, hook, "updating stale spending objects")

		updatedSpending1 := testutils.MustRetrieve(t, spending1)
		assert.Greater(t, updatedSpending1.NextRecurrence, spending1.NextRecurrence,
			"first spending next recurrence should have been advanced")

		updatedSpending2 := testutils.MustRetrieve(t, spending2)
		assert.Greater(t, updatedSpending2.NextRecurrence, spending2.NextRecurrence,
			"second spending next recurrence should have been advanced")
	})

	t.Run("spending with different funding schedules", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		timezone := testutils.MustEz(t, user.Account.GetTimezone)

		fundingRule1 := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", clock.Now())
		fundingSchedule1 := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday 1",
			Description:            "First payday",
			RuleSet:                fundingRule1,
			NextRecurrence:         fundingRule1.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule1.After(clock.Now(), false),
		})

		fundingRule2 := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=-1", clock.Now())
		fundingSchedule2 := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday 2",
			Description:            "Second payday",
			RuleSet:                fundingRule2,
			NextRecurrence:         fundingRule2.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule2.After(clock.Now(), false),
		})

		spendingRule1 := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule1.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending1 := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule1.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Expense for Schedule 1",
			Description:       "Linked to first funding schedule",
			TargetAmount:      5000,
			CurrentAmount:     5000,
			RuleSet:           spendingRule1,
			NextRecurrence:    spendingRule1.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		spendingRule2 := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=FR", clock.Now())
		spendingRule2.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending2 := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule2.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Expense for Schedule 2",
			Description:       "Linked to second funding schedule",
			TargetAmount:      3000,
			CurrentAmount:     3000,
			RuleSet:           spendingRule2,
			NextRecurrence:    spendingRule2.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending1.AccountId,
					BankAccountId: spending1.BankAccountId,
				},
			)
			assert.NoError(t, err, "should process spending with different funding schedules")
		}

		// Verify both spending objects were updated even though they reference
		// different funding schedules.
		updatedSpending1 := testutils.MustRetrieve(t, spending1)
		assert.Greater(t, updatedSpending1.NextRecurrence, spending1.NextRecurrence,
			"first spending next recurrence should have been advanced")

		updatedSpending2 := testutils.MustRetrieve(t, spending2)
		assert.Greater(t, updatedSpending2.NextRecurrence, spending2.NextRecurrence,
			"second spending next recurrence should have been advanced")
	})

	t.Run("goal spending advances recurrence", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
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

		// Create a goal with a stale next recurrence. Goals use
		// CurrentAmount + UsedAmount as their progress amount.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=YEARLY;INTERVAL=1;BYMONTH=1;BYMONTHDAY=1", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-366 * 24 * time.Hour))
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeGoal,
			Name:              "Savings Goal",
			Description:       "Stale goal",
			TargetAmount:      100000,
			CurrentAmount:     20000,
			UsedAmount:        5000,
			RuleSet:           spendingRule,
			NextRecurrence:    spendingRule.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     spending.AccountId,
					BankAccountId: spending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should process stale goal spending")
		}

		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.Greater(t, updatedSpending.NextRecurrence, spending.NextRecurrence,
			"goal next recurrence should have been advanced")
		assert.NotNil(t, updatedSpending.LastRecurrence,
			"last recurrence should be set for the goal")
	})

	t.Run("mix of stale and non-stale spending", func(t *testing.T) {
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

		// Stale spending - should be processed.
		staleRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		staleRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		staleSpending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Stale Expense",
			Description:       "Should be processed",
			TargetAmount:      5000,
			CurrentAmount:     5000,
			RuleSet:           staleRule,
			NextRecurrence:    staleRule.Before(clock.Now(), true),
			CreatedAt:         clock.Now(),
		})

		// Non-stale spending - should be left alone.
		freshRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", clock.Now())
		freshSpending := testutils.MustInsert(t, models.Spending{
			AccountId:         bankAccount.AccountId,
			BankAccountId:     bankAccount.BankAccountId,
			FundingScheduleId: fundingSchedule.FundingScheduleId,
			SpendingType:      models.SpendingTypeExpense,
			Name:              "Fresh Expense",
			Description:       "Should not be processed",
			TargetAmount:      10000,
			CurrentAmount:     0,
			RuleSet:           freshRule,
			NextRecurrence:    freshRule.After(clock.Now(), false),
			CreatedAt:         clock.Now(),
		})

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := spending_jobs.ProcessSpending(
				mockqueue.NewMockContext(context),
				spending_jobs.ProcessSpendingArguments{
					AccountId:     staleSpending.AccountId,
					BankAccountId: staleSpending.BankAccountId,
				},
			)
			assert.NoError(t, err, "should process mix of stale and non-stale spending")
		}

		testutils.MustHaveLogMessage(t, hook, "updating stale spending objects")

		// Verify the stale spending was updated.
		updatedStale := testutils.MustRetrieve(t, staleSpending)
		assert.Greater(t, updatedStale.NextRecurrence, staleSpending.NextRecurrence,
			"stale spending next recurrence should have been advanced")

		// Verify the non-stale spending was not touched.
		updatedFresh := testutils.MustRetrieve(t, freshSpending)
		assert.Equal(t, freshSpending.NextRecurrence.UTC(), updatedFresh.NextRecurrence.UTC(),
			"non-stale spending next recurrence should not have changed")
		assert.EqualValues(t, int64(0), updatedFresh.CurrentAmount,
			"non-stale spending current amount should not have changed")
	})
}
