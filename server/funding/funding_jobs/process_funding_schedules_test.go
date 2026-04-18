package funding_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/funding/funding_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProcessFundingSchedulesCron(t *testing.T) {
	t.Run("no funding schedules to process", func(t *testing.T) {
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

			err := funding_jobs.ProcessFundingSchedulesCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("enqueues funding schedules that are due", func(t *testing.T) {
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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		// Move time forward past the next recurrence so the funding schedule is
		// due for processing.
		clock.Add(32 * 24 * time.Hour)

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(funding_jobs.ProcessFundingSchedule),
				gomock.Any(),
				gomock.Eq(funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
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

			err := funding_jobs.ProcessFundingSchedulesCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})
}

func TestProcessFundingSchedule(t *testing.T) {
	t.Run("happy path with spending", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", clock.Now())
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Test Expense",
			Description:            "A test expense",
			TargetAmount:           10000,
			CurrentAmount:          0,
			NextContributionAmount: 5000,
			Ruleset:                spendingRule,
			NextRecurrence:         spendingRule.After(clock.Now(), false),
			CreatedAt:              clock.Now(),
		})

		// Move time forward past the funding schedule's next recurrence so it is
		// due for processing.
		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should process funding schedule successfully")
		}

		// Verify the spending was updated with the contribution.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(5000), updatedSpending.CurrentAmount,
			"current amount should reflect the contribution")

		// Verify the funding schedule's next recurrence was updated.
		updatedFundingSchedule := testutils.MustRetrieve(t, fundingSchedule)
		assert.Greater(t, updatedFundingSchedule.NextRecurrence, fundingSchedule.NextRecurrence,
			"next recurrence should have been advanced")
	})

	t.Run("no spending objects for funding schedule", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		// Move time forward past the funding schedule's next recurrence.
		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should succeed even with no spending objects")
		}

		// Verify the funding schedule's next recurrence was still updated.
		updatedFundingSchedule := testutils.MustRetrieve(t, fundingSchedule)
		assert.Greater(t, updatedFundingSchedule.NextRecurrence, fundingSchedule.NextRecurrence,
			"next recurrence should have been advanced")
	})

	t.Run("paused spending is skipped", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", clock.Now())
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Paused Expense",
			Description:            "This expense is paused",
			TargetAmount:           10000,
			CurrentAmount:          0,
			NextContributionAmount: 5000,
			IsPaused:               true,
			Ruleset:                spendingRule,
			NextRecurrence:         spendingRule.After(clock.Now(), false),
			CreatedAt:              clock.Now(),
		})

		// Move time forward past the funding schedule's next recurrence.
		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should succeed with paused spending")
		}

		// Verify the spending was NOT updated because it is paused.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(0), updatedSpending.CurrentAmount,
			"current amount should not change for paused spending")

		// Verify the funding schedule's next recurrence was still advanced even
		// though the spending was paused.
		updatedFundingSchedule := testutils.MustRetrieve(t, fundingSchedule)
		assert.Greater(t, updatedFundingSchedule.NextRecurrence, fundingSchedule.NextRecurrence,
			"next recurrence should have been advanced")
	})

	t.Run("spending already at target amount is skipped", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", clock.Now())
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Fully Funded Expense",
			Description:            "Already at target",
			TargetAmount:           10000,
			CurrentAmount:          10000,
			NextContributionAmount: 0,
			Ruleset:                spendingRule,
			NextRecurrence:         spendingRule.After(clock.Now(), false),
			CreatedAt:              clock.Now(),
		})

		// Move time forward past the funding schedule's next recurrence.
		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should succeed with fully funded spending")
		}

		// Verify the spending current amount was not modified.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(10000), updatedSpending.CurrentAmount,
			"current amount should not change for already funded spending")

		// Verify the funding schedule's next recurrence was still advanced even
		// though no spending was updated.
		updatedFundingSchedule := testutils.MustRetrieve(t, fundingSchedule)
		assert.Greater(t, updatedFundingSchedule.NextRecurrence, fundingSchedule.NextRecurrence,
			"next recurrence should have been advanced")
	})

	t.Run("funding schedule not yet due is skipped", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", clock.Now())
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Test Expense",
			Description:            "Should not be funded yet",
			TargetAmount:           10000,
			CurrentAmount:          0,
			NextContributionAmount: 5000,
			Ruleset:                spendingRule,
			NextRecurrence:         spendingRule.After(clock.Now(), false),
			CreatedAt:              clock.Now(),
		})

		// Do NOT move the clock forward. The funding schedule's next recurrence
		// is still in the future, so CalculateNextOccurrence should return false
		// and the schedule should be skipped entirely.

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should succeed even when funding schedule is not yet due")
		}

		// Verify the spending was not touched.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(0), updatedSpending.CurrentAmount,
			"current amount should remain unchanged")

		// Verify the funding schedule's next recurrence was NOT changed since it
		// was not due yet.
		updatedFundingSchedule := testutils.MustRetrieve(t, fundingSchedule)
		assert.Equal(t, fundingSchedule.NextRecurrence, updatedFundingSchedule.NextRecurrence,
			"next recurrence should not have changed")
	})

	t.Run("goal spending type uses progress amount", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		// Create a goal where CurrentAmount + UsedAmount already meets the
		// target. For goals, GetProgressAmount returns CurrentAmount + UsedAmount,
		// so with CurrentAmount=3000 and UsedAmount=7000 the progress is 10000
		// which equals the target. This spending should be skipped.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=YEARLY;INTERVAL=1;BYMONTH=12;BYMONTHDAY=25", clock.Now())
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeGoal,
			Name:                   "Savings Goal",
			Description:            "Goal with used amount",
			TargetAmount:           10000,
			CurrentAmount:          3000,
			UsedAmount:             7000,
			NextContributionAmount: 0,
			Ruleset:                spendingRule,
			NextRecurrence:         spendingRule.After(clock.Now(), false),
			CreatedAt:              clock.Now(),
		})

		// Move time forward past the funding schedule's next recurrence.
		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should succeed with goal at target via used amount")
		}

		// Verify the spending was not funded further since the goal's progress
		// (CurrentAmount + UsedAmount) already meets the target.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(3000), updatedSpending.CurrentAmount,
			"current amount should not change for goal already at target")
	})

	t.Run("goal spending type receives contribution", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		// Create a goal that still needs funding. CurrentAmount=2000,
		// UsedAmount=3000, so progress is 5000 which is less than the target of
		// 10000.
		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=YEARLY;INTERVAL=1;BYMONTH=12;BYMONTHDAY=25", clock.Now())
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule.FundingScheduleId,
			SpendingType:           models.SpendingTypeGoal,
			Name:                   "Savings Goal",
			Description:            "Goal needing funding",
			TargetAmount:           10000,
			CurrentAmount:          2000,
			UsedAmount:             3000,
			NextContributionAmount: 1000,
			Ruleset:                spendingRule,
			NextRecurrence:         spendingRule.After(clock.Now(), false),
			CreatedAt:              clock.Now(),
		})

		// Move time forward past the funding schedule's next recurrence.
		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should process goal spending successfully")
		}

		// Verify the goal received its contribution.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(3000), updatedSpending.CurrentAmount,
			"current amount should have increased by the contribution amount")
	})

	t.Run("multiple funding schedules for same bank account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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

		fundingRule1 := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", clock.Now())
		fundingSchedule1 := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday 1",
			Description:            "First payday",
			Ruleset:                fundingRule1,
			NextRecurrence:         fundingRule1.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule1.After(clock.Now(), false),
		})

		fundingRule2 := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=-1", clock.Now())
		fundingSchedule2 := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday 2",
			Description:            "Second payday",
			Ruleset:                fundingRule2,
			NextRecurrence:         fundingRule2.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule2.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", clock.Now())
		spending1 := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule1.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Expense for Schedule 1",
			Description:            "Linked to first funding schedule",
			TargetAmount:           10000,
			CurrentAmount:          0,
			NextContributionAmount: 5000,
			Ruleset:                spendingRule,
			NextRecurrence:         spendingRule.After(clock.Now(), false),
			CreatedAt:              clock.Now(),
		})

		spendingRule2 := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1", clock.Now())
		spending2 := testutils.MustInsert(t, models.Spending{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			FundingScheduleId:      fundingSchedule2.FundingScheduleId,
			SpendingType:           models.SpendingTypeExpense,
			Name:                   "Expense for Schedule 2",
			Description:            "Linked to second funding schedule",
			TargetAmount:           8000,
			CurrentAmount:          0,
			NextContributionAmount: 4000,
			Ruleset:                spendingRule2,
			NextRecurrence:         spendingRule2.After(clock.Now(), false),
			CreatedAt:              clock.Now(),
		})

		// Move time forward past both funding schedules' next recurrences.
		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:     bankAccount.AccountId,
					BankAccountId: bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{
						fundingSchedule1.FundingScheduleId,
						fundingSchedule2.FundingScheduleId,
					},
				},
			)
			assert.NoError(t, err, "should process multiple funding schedules successfully")
		}

		// Verify both spending objects were updated.
		updatedSpending1 := testutils.MustRetrieve(t, spending1)
		assert.EqualValues(t, int64(5000), updatedSpending1.CurrentAmount,
			"first spending current amount should reflect its contribution")

		updatedSpending2 := testutils.MustRetrieve(t, spending2)
		assert.EqualValues(t, int64(4000), updatedSpending2.CurrentAmount,
			"second spending current amount should reflect its contribution")

		// Verify both funding schedules were advanced.
		updatedFundingSchedule1 := testutils.MustRetrieve(t, fundingSchedule1)
		assert.Greater(t, updatedFundingSchedule1.NextRecurrence, fundingSchedule1.NextRecurrence,
			"first funding schedule next recurrence should have been advanced")

		updatedFundingSchedule2 := testutils.MustRetrieve(t, fundingSchedule2)
		assert.Greater(t, updatedFundingSchedule2.NextRecurrence, fundingSchedule2.NextRecurrence,
			"second funding schedule next recurrence should have been advanced")
	})

	t.Run("auto creates deposit transaction on manual link", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		fundingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", clock.Now())
		estimatedDeposit := int64(100000)
		expectedFundingDate := fundingRule.After(clock.Now(), false)
		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday",
			Description:            "Payday",
			RuleSet:                fundingRule,
			NextRecurrence:         expectedFundingDate,
			NextRecurrenceOriginal: expectedFundingDate,
			EstimatedDeposit:       &estimatedDeposit,
			AutoCreateTransaction:  true,
		})

		// Move time forward past the funding schedule's next recurrence.
		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should process funding schedule successfully")
		}

		// Verify the bank account balance was incremented by the deposit.
		updatedBankAccount := testutils.MustRetrieve(t, bankAccount)
		assert.EqualValues(t, bankAccount.AvailableBalance+estimatedDeposit, updatedBankAccount.AvailableBalance,
			"available balance should have been incremented by the deposit")
		assert.EqualValues(t, bankAccount.CurrentBalance+estimatedDeposit, updatedBankAccount.CurrentBalance,
			"current balance should have been incremented by the deposit")

		// Verify a deposit transaction was created.
		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		transactions, err := repo.GetTransactions(t.Context(), bankAccount.BankAccountId, 100, 0)
		assert.NoError(t, err, "should retrieve transactions")
		assert.Len(t, transactions, 1, "exactly one deposit transaction should have been created")
		assert.EqualValues(t, -estimatedDeposit, transactions[0].Amount, "deposit amount should be negative")
		assert.Equal(t, models.TransactionSourceManual, transactions[0].Source, "transaction source should be manual")
		assert.Equal(t, fundingSchedule.Name, transactions[0].Name, "transaction name should match the funding schedule name")
		assert.Nil(t, transactions[0].SpendingId, "deposit should not be allocated to a spending")
		assert.True(t, transactions[0].Date.Equal(expectedFundingDate), "transaction date should match the funding date")
	})

	t.Run("does not auto create deposit on non-manual link", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

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
		estimatedDeposit := int64(100000)
		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday",
			Description:            "Payday",
			RuleSet:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
			EstimatedDeposit:       &estimatedDeposit,
			AutoCreateTransaction:  true,
		})

		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should process funding schedule successfully")
		}

		// Verify the bank account balance did not change on a non-manual link.
		updatedBankAccount := testutils.MustRetrieve(t, bankAccount)
		assert.EqualValues(t, bankAccount.AvailableBalance, updatedBankAccount.AvailableBalance, "available balance should not change")
		assert.EqualValues(t, bankAccount.CurrentBalance, updatedBankAccount.CurrentBalance, "current balance should not change")

		// Verify no transactions were created.
		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		transactions, err := repo.GetTransactions(t.Context(), bankAccount.BankAccountId, 100, 0)
		assert.NoError(t, err, "should retrieve transactions")
		assert.Empty(t, transactions, "no transactions should have been created on a non-manual link")
	})

	t.Run("does not auto create deposit when estimated deposit is nil", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
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
			EstimatedDeposit:       nil,
			AutoCreateTransaction:  true,
		})

		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should process funding schedule successfully")
		}

		// Verify the bank account balance did not change without an estimated deposit.
		updatedBankAccount := testutils.MustRetrieve(t, bankAccount)
		assert.EqualValues(t, bankAccount.AvailableBalance, updatedBankAccount.AvailableBalance, "available balance should not change")
		assert.EqualValues(t, bankAccount.CurrentBalance, updatedBankAccount.CurrentBalance, "current balance should not change")

		// Verify no transactions were created.
		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		transactions, err := repo.GetTransactions(t.Context(), bankAccount.BankAccountId, 100, 0)
		assert.NoError(t, err, "should retrieve transactions")
		assert.Empty(t, transactions, "no transactions should have been created without an estimated deposit")
	})

	t.Run("does not auto create deposit when flag is off", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		timezone := testutils.MustEz(t, user.Account.GetTimezone)
		fundingRule := testutils.RuleToSet(t, timezone, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", clock.Now())
		estimatedDeposit := int64(100000)
		fundingSchedule := testutils.MustInsert(t, models.FundingSchedule{
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			Name:                   "Payday",
			Description:            "Payday",
			RuleSet:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
			EstimatedDeposit:       &estimatedDeposit,
			AutoCreateTransaction:  false,
		})

		clock.Add(32 * 24 * time.Hour)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

			err := funding_jobs.ProcessFundingSchedule(
				mockqueue.NewMockContext(context),
				funding_jobs.ProcessFundingScheduleArguments{
					AccountId:          bankAccount.AccountId,
					BankAccountId:      bankAccount.BankAccountId,
					FundingScheduleIds: []models.ID[models.FundingSchedule]{fundingSchedule.FundingScheduleId},
				},
			)
			assert.NoError(t, err, "should process funding schedule successfully")
		}

		// Verify the bank account balance did not change when the flag is off.
		updatedBankAccount := testutils.MustRetrieve(t, bankAccount)
		assert.EqualValues(t, bankAccount.AvailableBalance, updatedBankAccount.AvailableBalance, "available balance should not change")
		assert.EqualValues(t, bankAccount.CurrentBalance, updatedBankAccount.CurrentBalance, "current balance should not change")

		// Verify no transactions were created.
		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		transactions, err := repo.GetTransactions(t.Context(), bankAccount.BankAccountId, 100, 0)
		assert.NoError(t, err, "should retrieve transactions")
		assert.Empty(t, transactions, "no transactions should have been created when the flag is off")
	})
}
