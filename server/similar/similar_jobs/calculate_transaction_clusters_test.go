package similar_jobs_test

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
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/similar/similar_jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// givenIHaveMonthlyTransactions creates a transaction with the same name on the
// same day of the month for the specified number of months.
func givenIHaveMonthlyTransactions(
	t *testing.T,
	clock clock.Clock,
	bankAccount models.BankAccount,
	name string,
	amount int64,
	months int,
) []models.Transaction {
	repo := repository.NewRepositoryFromSession(
		clock,
		bankAccount.Link.CreatedBy,
		bankAccount.AccountId,
		testutils.GetPgDatabase(t),
		testutils.GetLog(t),
	)

	transactions := make([]models.Transaction, months)
	for i := range transactions {
		transactions[i] = models.Transaction{
			AccountId:            bankAccount.AccountId,
			BankAccountId:        bankAccount.BankAccountId,
			Amount:               amount,
			Date:                 time.Date(2026, time.Month(i+1), 8, 0, 0, 0, 0, time.UTC),
			Name:                 name,
			OriginalName:         name,
			MerchantName:         name,
			OriginalMerchantName: name,
			Source:               models.TransactionSourceUpload,
			CreatedAt:            clock.Now(),
		}
		require.NoError(
			t,
			repo.CreateTransaction(t.Context(), bankAccount.BankAccountId, &transactions[i]),
			"must be able to seed transaction",
		)
	}

	return transactions
}

func TestCalculateTransactionClusters(t *testing.T) {
	t.Run("enqueues one recurring calculation for the bank account", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		givenIHaveMonthlyTransactions(t, clock, bankAccount, "GITHUB INC", 800, 6)
		givenIHaveMonthlyTransactions(t, clock, bankAccount, "SENTRY IO", 2600, 6)

		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(recurring_jobs.CalculateRecurringTransactions),
				gomock.Any(),
				gomock.Eq(recurring_jobs.CalculateRecurringTransactionsArguments{
					AccountId:     bankAccount.AccountId,
					BankAccountId: bankAccount.BankAccountId,
				}),
			).
			Return(nil).
			Times(1)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := similar_jobs.CalculateTransactionClusters(
			mockqueue.NewMockContext(context),
			similar_jobs.CalculateTransactionClustersArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
			},
		)
		assert.NoError(t, err, "must be able to calculate transaction clusters")

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)
		clusterIds, err := repo.GetTransactionClusterIds(t.Context(), bankAccount.BankAccountId)
		require.NoError(t, err, "must be able to read the transaction clusters")
		assert.Len(t, clusterIds, 2, "there should be a cluster for each merchant")
	})

	t.Run("does not enqueue without any clusters", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		enqueuer.EXPECT().EnqueueAt(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := similar_jobs.CalculateTransactionClusters(
			mockqueue.NewMockContext(context),
			similar_jobs.CalculateTransactionClustersArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
			},
		)
		assert.NoError(t, err, "must be able to calculate transaction clusters")
	})
}
