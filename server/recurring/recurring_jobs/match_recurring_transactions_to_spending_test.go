package recurring_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/recurring/recurring_jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"go.uber.org/mock/gomock"
)

// givenIHaveSpending will create a spending object of the specified type with a
// funding schedule of its own.
func givenIHaveSpending(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
	spendingType models.SpendingType,
	name string,
) models.Spending {
	fundingSchedule := fixtures.GivenIHaveAFundingSchedule(t, clock, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15", false)
	spendingRule := testutils.Must(t, models.NewRuleSet, "DTSTART:20260101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1")
	return testutils.MustInsert(t, models.Spending{
		AccountId:         bankAccount.AccountId,
		BankAccountId:     bankAccount.BankAccountId,
		FundingScheduleId: fundingSchedule.FundingScheduleId,
		SpendingType:      spendingType,
		Name:              name,
		TargetAmount:      800,
		RuleSet:           spendingRule,
		NextRecurrence:    spendingRule.After(clock.Now(), false),
		CreatedAt:         clock.Now(),
	})
}

// givenTheTransactionsWereSpentFrom will set the spending on the provided
// transactions. Only the spending column is written so the recurring ID the
// job assigned to each transaction is left alone.
func givenTheTransactionsWereSpentFrom(
	t *testing.T,
	spending models.Spending,
	transactions ...models.Transaction,
) {
	ids := make([]models.ID[models.Transaction], len(transactions))
	for i := range transactions {
		ids[i] = transactions[i].TransactionId
	}

	_, err := testutils.GetPgDatabase(t).NewUpdate().
		Model(new(models.Transaction)).
		Set(`"spending_id" = ?`, spending.SpendingId).
		Where(`"transaction"."account_id" = ?`, spending.AccountId).
		Where(`"transaction"."transaction_id" IN (?)`, bun.List(ids)).
		Exec(t.Context())
	require.NoError(t, err, "must be able to set the spending on the transactions")
}

// givenIHaveARecurringExpense will create six monthly charges in a single
// cluster and calculate the recurring transaction for them. The charges are
// returned oldest first.
func givenIHaveARecurringExpense(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
) (models.TransactionCluster, []models.Transaction) {
	cluster, transactions := givenIHaveAClusterWithTransactions(
		t,
		clock,
		bankAccount,
		repeatAmount(800, 6),
		monthlyDates(accountTimezone(t, bankAccount), 6),
	)
	require.NoError(t, runCalculateRecurringTransactions(t, clock, bankAccount), "must calculate recurring transactions")
	require.Len(t, readRecurringByCluster(t, clock, cluster), 1, "should have a single recurring transaction")

	return cluster, transactions
}

// runMatchRecurringTransactionsToSpending will run the job for the whole
// bank account.
func runMatchRecurringTransactionsToSpending(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
) error {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := mockgen.NewMockContext(ctrl)
	context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
	context.EXPECT().Clock().Return(clock).AnyTimes()
	context.EXPECT().DB().Return(testutils.GetPgDatabase(t)).AnyTimes()
	context.EXPECT().Log().Return(testutils.GetLog(t)).AnyTimes()

	return recurring_jobs.MatchRecurringTransactionsToSpending(
		mockqueue.NewMockContext(context),
		recurring_jobs.MatchRecurringTransactionsToSpendingArguments{
			AccountId:     bankAccount.AccountId,
			BankAccountId: bankAccount.BankAccountId,
		},
	)
}

func TestMatchRecurringTransactionsToSpending(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		expense := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Github")
		cluster, transactions := givenIHaveARecurringExpense(t, clock, bankAccount)
		// Two of the three most recent charges were spent from the expense.
		givenTheTransactionsWereSpentFrom(t, expense, transactions[3], transactions[5])

		require.NoError(t, runMatchRecurringTransactionsToSpending(t, clock, bankAccount), "job must succeed")
		result := readRecurringByCluster(t, clock, cluster)
		require.Len(t, result, 1, "should still have a single recurring transaction")
		require.NotNil(t, result[0].SpendingId, "should have been linked to the expense")
		assert.Equal(t, expense.SpendingId, *result[0].SpendingId, "should be linked to the expense it was spent from")
		assert.Nil(t, result[0].FundingScheduleId, "should not be linked to a funding schedule")
		assert.True(t, result[0].AutoMatched, "should be marked as auto matched")
	})

	t.Run("one recent transaction", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		expense := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Github")
		cluster, transactions := givenIHaveARecurringExpense(t, clock, bankAccount)
		// The older charges were all spent from the expense, but only one of the
		// three most recent ones was.
		givenTheTransactionsWereSpentFrom(
			t,
			expense,
			transactions[0],
			transactions[1],
			transactions[2],
			transactions[5],
		)

		require.NoError(t, runMatchRecurringTransactionsToSpending(t, clock, bankAccount), "job must succeed")
		result := readRecurringByCluster(t, clock, cluster)
		require.Len(t, result, 1, "should still have a single recurring transaction")
		assert.Nil(t, result[0].SpendingId, "should not have been linked to anything")
	})

	t.Run("only members count", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		expense := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Github")
		cluster, _ := givenIHaveARecurringExpense(t, clock, bankAccount)

		// Newer one off charges in the same cluster, like random Amazon orders,
		// that were spent from the expense but are not part of the recurring
		// transaction.
		timezone := accountTimezone(t, bankAccount)
		for _, day := range []int{20, 25, 28} {
			transaction := testutils.MustInsert(t, models.Transaction{
				AccountId:            bankAccount.AccountId,
				BankAccountId:        bankAccount.BankAccountId,
				TransactionClusterId: &cluster.TransactionClusterId,
				SpendingId:           &expense.SpendingId,
				Amount:               1234,
				Date:                 time.Date(2026, 6, day, 0, 0, 0, 0, timezone),
				Name:                 "GITHUB INC",
				OriginalName:         "GITHUB INC",
				MerchantName:         "GITHUB INC",
				OriginalMerchantName: "GITHUB INC",
				Source:               models.TransactionSourceUpload,
				CreatedAt:            clock.Now(),
			})
			require.Nil(t, transaction.TransactionRecurringId, "one off charge must not be part of the recurring transaction")
		}

		require.NoError(t, runMatchRecurringTransactionsToSpending(t, clock, bankAccount), "job must succeed")
		result := readRecurringByCluster(t, clock, cluster)
		require.Len(t, result, 1, "should still have a single recurring transaction")
		assert.Nil(t, result[0].SpendingId, "should not have been linked to anything")
	})

	t.Run("does not link goals", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		goal := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeGoal, "Github")
		cluster, transactions := givenIHaveARecurringExpense(t, clock, bankAccount)
		givenTheTransactionsWereSpentFrom(t, goal, transactions[3], transactions[4], transactions[5])

		require.NoError(t, runMatchRecurringTransactionsToSpending(t, clock, bankAccount), "job must succeed")
		result := readRecurringByCluster(t, clock, cluster)
		require.Len(t, result, 1, "should still have a single recurring transaction")
		assert.Nil(t, result[0].SpendingId, "should not have been linked to the goal")
	})

	t.Run("existing link", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		linked := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Github")
		other := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Software")
		cluster, transactions := givenIHaveARecurringExpense(t, clock, bankAccount)
		recurring := readRecurringByCluster(t, clock, cluster)
		recurring[0].SpendingId = &linked.SpendingId
		testutils.MustDBUpdate(t, &recurring[0])
		givenTheTransactionsWereSpentFrom(t, other, transactions[3], transactions[4], transactions[5])

		require.NoError(t, runMatchRecurringTransactionsToSpending(t, clock, bankAccount), "job must succeed")
		result := readRecurringByCluster(t, clock, cluster)
		require.Len(t, result, 1, "should still have a single recurring transaction")
		require.NotNil(t, result[0].SpendingId, "should still be linked")
		assert.Equal(t, linked.SpendingId, *result[0].SpendingId, "should still be linked to the original expense")
		assert.False(t, result[0].AutoMatched, "a link the user made must not be marked as auto matched")
	})

	t.Run("expense already linked", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		expense := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Github")
		linkedCluster, _ := givenIHaveARecurringExpense(t, clock, bankAccount)
		linked := readRecurringByCluster(t, clock, linkedCluster)
		linked[0].SpendingId = &expense.SpendingId
		testutils.MustDBUpdate(t, &linked[0])
		otherCluster, transactions := givenIHaveARecurringExpense(t, clock, bankAccount)
		givenTheTransactionsWereSpentFrom(t, expense, transactions[3], transactions[4], transactions[5])

		// An expense can only be linked to one recurring transaction, linking it
		// again would fail the whole job.
		require.NoError(t, runMatchRecurringTransactionsToSpending(t, clock, bankAccount), "job must succeed")
		result := readRecurringByCluster(t, clock, otherCluster)
		require.Len(t, result, 1, "should still have a single recurring transaction")
		assert.Nil(t, result[0].SpendingId, "should not take the expense from the other recurring transaction")
		assert.False(t, result[0].AutoMatched, "should not be marked as auto matched")
	})

	t.Run("no recurring transactions", func(t *testing.T) {
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC))
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		assert.NoError(t, runMatchRecurringTransactionsToSpending(t, clock, bankAccount), "job must succeed")
	})
}
