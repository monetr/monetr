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
	"github.com/monetr/monetr/server/repository"
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
			Ruleset:                fundingRule,
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
			Ruleset:           spendingRule,
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

	t.Run("auto creates transaction for stale expense on manual link", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
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

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour)) // Allow past times.
		expectedDueDate := spendingRule.Before(clock.Now(), true)
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:             bankAccount.AccountId,
			BankAccountId:         bankAccount.BankAccountId,
			FundingScheduleId:     fundingSchedule.FundingScheduleId,
			SpendingType:          models.SpendingTypeExpense,
			Name:                  "Auto Expense",
			Description:           "Auto expense",
			TargetAmount:          5000,
			CurrentAmount:         5000,
			Ruleset:               spendingRule,
			NextRecurrence:        expectedDueDate, // Make it so it recurs next in the past. (STALE)
			AutoCreateTransaction: true,
			CreatedAt:             clock.Now(),
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
			assert.NoError(t, err, "should run job successfully")
		}

		// Verify the expense was allocated to the new transaction.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(0), updatedSpending.CurrentAmount, "current amount should have been deducted")

		// Verify the bank account balance was decremented by the target amount.
		updatedBankAccount := testutils.MustRetrieve(t, bankAccount)
		assert.EqualValues(t, bankAccount.AvailableBalance-5000, updatedBankAccount.AvailableBalance, "available balance should have been decremented")
		assert.EqualValues(t, bankAccount.CurrentBalance-5000, updatedBankAccount.CurrentBalance, "current balance should have been decremented")

		// Verify a transaction was created for the expense.
		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		transactions, err := repo.GetTransactions(t.Context(), bankAccount.BankAccountId, 100, 0)
		assert.NoError(t, err, "should retrieve transactions")
		assert.Len(t, transactions, 1, "exactly one transaction should have been created")
		assert.EqualValues(t, int64(5000), transactions[0].Amount, "transaction amount should match the expense target")
		assert.Equal(t, models.TransactionSourceManual, transactions[0].Source, "transaction source should be manual")
		assert.Equal(t, spending.Name, transactions[0].Name, "transaction name should match the expense name")
		assert.NotNil(t, transactions[0].SpendingId, "transaction should be allocated to a spending")
		assert.Equal(t, spending.SpendingId, *transactions[0].SpendingId, "transaction should be allocated to the auto-create expense")
		assert.True(t, transactions[0].Date.Equal(expectedDueDate), "transaction date should match the expense due date")
	})

	t.Run("does not auto create transaction on non-manual link", func(t *testing.T) {
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
			Ruleset:                fundingRule,
			NextRecurrence:         fundingRule.After(clock.Now(), false),
			NextRecurrenceOriginal: fundingRule.After(clock.Now(), false),
		})

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:             bankAccount.AccountId,
			BankAccountId:         bankAccount.BankAccountId,
			FundingScheduleId:     fundingSchedule.FundingScheduleId,
			SpendingType:          models.SpendingTypeExpense,
			Name:                  "Plaid Expense",
			Description:           "Plaid expense",
			TargetAmount:          5000,
			CurrentAmount:         5000,
			Ruleset:               spendingRule,
			NextRecurrence:        spendingRule.Before(clock.Now(), true),
			AutoCreateTransaction: true,
			CreatedAt:             clock.Now(),
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
			assert.NoError(t, err, "should run job successfully")
		}

		// Verify the expense was not deducted from on a non-manual link.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(5000), updatedSpending.CurrentAmount, "current amount should not change on a non-manual link")

		// Verify the bank account balance did not change.
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

	t.Run("does not auto create transaction for goal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
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

		// Goals do not use a recurrence rule, but they have a single NextRecurrence
		// in the past to make them stale.
		goal := testutils.MustInsert(t, models.Spending{
			AccountId:             bankAccount.AccountId,
			BankAccountId:         bankAccount.BankAccountId,
			FundingScheduleId:     fundingSchedule.FundingScheduleId,
			SpendingType:          models.SpendingTypeGoal,
			Name:                  "Auto Goal",
			Description:           "Auto goal",
			TargetAmount:          5000,
			CurrentAmount:         5000,
			NextRecurrence:        clock.Now().Add(-24 * time.Hour),
			AutoCreateTransaction: true, // Defensive: jobs must skip goals even if the flag was set somehow.
			CreatedAt:             clock.Now(),
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
					AccountId:     goal.AccountId,
					BankAccountId: goal.BankAccountId,
				},
			)
			assert.NoError(t, err, "should run job successfully")
		}

		// Verify the goal was not deducted from.
		updatedGoal := testutils.MustRetrieve(t, goal)
		assert.EqualValues(t, int64(5000), updatedGoal.CurrentAmount, "current amount should not change for a goal")

		// Verify the bank account balance did not change.
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
		assert.Empty(t, transactions, "no transactions should have been created for a goal")
	})

	t.Run("does not auto create transaction when flag is off", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
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

		spendingRule := testutils.RuleToSet(t, timezone, "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO", clock.Now())
		spendingRule.DTStart(clock.Now().Add(-8 * 24 * time.Hour))
		spending := testutils.MustInsert(t, models.Spending{
			AccountId:             bankAccount.AccountId,
			BankAccountId:         bankAccount.BankAccountId,
			FundingScheduleId:     fundingSchedule.FundingScheduleId,
			SpendingType:          models.SpendingTypeExpense,
			Name:                  "Manual Expense",
			Description:           "Manual expense without auto create",
			TargetAmount:          5000,
			CurrentAmount:         5000,
			Ruleset:               spendingRule,
			NextRecurrence:        spendingRule.Before(clock.Now(), true),
			AutoCreateTransaction: false,
			CreatedAt:             clock.Now(),
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
			assert.NoError(t, err, "should run job successfully")
		}

		// Verify the expense was not deducted from.
		updatedSpending := testutils.MustRetrieve(t, spending)
		assert.EqualValues(t, int64(5000), updatedSpending.CurrentAmount, "current amount should not change when flag is off")

		// Verify the bank account balance did not change.
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
		assert.Empty(t, transactions, "no transactions should have been created when flag is off")
	})
}
