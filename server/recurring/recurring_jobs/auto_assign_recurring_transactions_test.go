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
	"go.uber.org/mock/gomock"
)

// runAutoAssignRecurringTransactions will run the job for the provided
// transactions.
func runAutoAssignRecurringTransactions(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
	transactionIds ...models.ID[models.Transaction],
) error {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	context := mockgen.NewMockContext(ctrl)
	context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
	context.EXPECT().Clock().Return(clock).AnyTimes()
	context.EXPECT().DB().Return(testutils.GetPgDatabase(t)).AnyTimes()
	context.EXPECT().Log().Return(testutils.GetLog(t)).AnyTimes()

	return recurring_jobs.AutoAssignRecurringTransactions(
		mockqueue.NewMockContext(context),
		recurring_jobs.AutoAssignRecurringTransactionsArguments{
			AccountId:      bankAccount.AccountId,
			BankAccountId:  bankAccount.BankAccountId,
			TransactionIds: transactionIds,
		},
	)
}

func TestAutoAssignRecurringTransactions(t *testing.T) {
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
		expense.CurrentAmount = 800
		testutils.MustDBUpdate(t, &expense)
		cluster, _ := givenIHaveARecurringExpense(t, clock, bankAccount)
		recurring := readRecurringByCluster(t, clock, cluster)[0]
		recurring.SpendingId = &expense.SpendingId
		recurring.AutoAssign = true
		testutils.MustDBUpdate(t, &recurring)
		clock.Add(8 * 24 * time.Hour)
		transaction := testutils.MustInsert(t, models.Transaction{
			TransactionId:          models.NewID[models.Transaction](),
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			TransactionClusterId:   &cluster.TransactionClusterId,
			TransactionRecurringId: &recurring.TransactionRecurringId,
			Amount:                 800,
			Date:                   time.Date(2026, 7, 8, 0, 0, 0, 0, accountTimezone(t, bankAccount)),
			Name:                   "GITHUB INC",
			OriginalName:           "GITHUB INC",
			MerchantName:           "GITHUB INC",
			OriginalMerchantName:   "GITHUB INC",
			Source:                 models.TransactionSourceUpload,
			CreatedAt:              clock.Now(),
		})

		err := runAutoAssignRecurringTransactions(t, clock, bankAccount, transaction.TransactionId)
		require.NoError(t, err, "job must succeed")

		stored := testutils.MustDBRead(t, transaction)
		if assert.NotNil(t, stored.SpendingId, "should have been spent from the expense") {
			assert.Equal(t, expense.SpendingId, *stored.SpendingId, "should be spent from the linked expense")
		}
		if assert.NotNil(t, stored.SpendingAmount, "should have a spending amount") {
			assert.EqualValues(t, 800, *stored.SpendingAmount, "should have spent the whole transaction")
		}
		assert.EqualValues(t, 0, testutils.MustDBRead(t, expense).CurrentAmount, "the expense should have been used up")
	})

	t.Run("already spent", func(t *testing.T) {
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
		expense.CurrentAmount = 800
		testutils.MustDBUpdate(t, &expense)
		other := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Something Else")
		cluster, _ := givenIHaveARecurringExpense(t, clock, bankAccount)
		recurring := readRecurringByCluster(t, clock, cluster)[0]
		recurring.SpendingId = &expense.SpendingId
		recurring.AutoAssign = true
		testutils.MustDBUpdate(t, &recurring)
		clock.Add(8 * 24 * time.Hour)
		transaction := testutils.MustInsert(t, models.Transaction{
			TransactionId:          models.NewID[models.Transaction](),
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			TransactionClusterId:   &cluster.TransactionClusterId,
			TransactionRecurringId: &recurring.TransactionRecurringId,
			Amount:                 800,
			Date:                   time.Date(2026, 7, 8, 0, 0, 0, 0, accountTimezone(t, bankAccount)),
			Name:                   "GITHUB INC",
			OriginalName:           "GITHUB INC",
			MerchantName:           "GITHUB INC",
			OriginalMerchantName:   "GITHUB INC",
			Source:                 models.TransactionSourceUpload,
			CreatedAt:              clock.Now(),
		})
		givenTheTransactionsWereSpentFrom(t, other, transaction)

		err := runAutoAssignRecurringTransactions(t, clock, bankAccount, transaction.TransactionId)
		require.NoError(t, err, "job must succeed")

		stored := testutils.MustDBRead(t, transaction)
		if assert.NotNil(t, stored.SpendingId, "should still be spent from something") {
			assert.Equal(t, other.SpendingId, *stored.SpendingId, "should still be spent from what it was before")
		}
		assert.EqualValues(t, 800, testutils.MustDBRead(t, expense).CurrentAmount, "the linked expense should not have been touched")
	})

	t.Run("before the spending existed", func(t *testing.T) {
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
		// The expense is created July 1st, the recurring charges are all on the 8th
		// of January through June.
		expense := givenIHaveSpending(t, clock, bankAccount, models.SpendingTypeExpense, "Github")
		expense.CurrentAmount = 800
		testutils.MustDBUpdate(t, &expense)
		cluster, transactions := givenIHaveARecurringExpense(t, clock, bankAccount)
		recurring := readRecurringByCluster(t, clock, cluster)[0]
		recurring.SpendingId = &expense.SpendingId
		recurring.AutoAssign = true
		testutils.MustDBUpdate(t, &recurring)

		err := runAutoAssignRecurringTransactions(t, clock, bankAccount, transactions[5].TransactionId)
		require.NoError(t, err, "job must succeed")

		stored := testutils.MustDBRead(t, transactions[5])
		assert.Nil(t, stored.SpendingId, "should not be spent from an expense that didnt exist yet")
		assert.EqualValues(t, 800, testutils.MustDBRead(t, expense).CurrentAmount, "the expense should not have been touched")
	})

	t.Run("auto assign is off", func(t *testing.T) {
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
		expense.CurrentAmount = 800
		testutils.MustDBUpdate(t, &expense)
		cluster, _ := givenIHaveARecurringExpense(t, clock, bankAccount)
		recurring := readRecurringByCluster(t, clock, cluster)[0]
		recurring.SpendingId = &expense.SpendingId
		testutils.MustDBUpdate(t, &recurring)
		clock.Add(8 * 24 * time.Hour)
		transaction := testutils.MustInsert(t, models.Transaction{
			TransactionId:          models.NewID[models.Transaction](),
			AccountId:              bankAccount.AccountId,
			BankAccountId:          bankAccount.BankAccountId,
			TransactionClusterId:   &cluster.TransactionClusterId,
			TransactionRecurringId: &recurring.TransactionRecurringId,
			Amount:                 800,
			Date:                   time.Date(2026, 7, 8, 0, 0, 0, 0, accountTimezone(t, bankAccount)),
			Name:                   "GITHUB INC",
			OriginalName:           "GITHUB INC",
			MerchantName:           "GITHUB INC",
			OriginalMerchantName:   "GITHUB INC",
			Source:                 models.TransactionSourceUpload,
			CreatedAt:              clock.Now(),
		})

		err := runAutoAssignRecurringTransactions(t, clock, bankAccount, transaction.TransactionId)
		require.NoError(t, err, "job must succeed")

		stored := testutils.MustDBRead(t, transaction)
		assert.Nil(t, stored.SpendingId, "should not be spent when auto assign is off")
		assert.EqualValues(t, 800, testutils.MustDBRead(t, expense).CurrentAmount, "the expense should not have been touched")
	})
}
