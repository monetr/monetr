package similar_jobs_test

import (
	"context"
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
	"github.com/uptrace/bun"
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

// recurringArgumentsForClusters returns the recurring job arguments that should
// be enqueued for every cluster that currently exists for the bank account.
func recurringArgumentsForClusters(
	t *testing.T,
	bankAccount models.BankAccount,
) []any {
	var clusters []models.TransactionCluster
	err := testutils.GetPgDatabase(t).NewSelect().
		Model(&clusters).
		Where(`"transaction_cluster"."account_id" = ?`, bankAccount.AccountId).
		Where(`"transaction_cluster"."bank_account_id" = ?`, bankAccount.BankAccountId).
		Scan(t.Context())
	require.NoError(t, err, "must be able to read the transaction clusters")

	result := make([]any, len(clusters))
	for i, cluster := range clusters {
		result[i] = recurring_jobs.CalculateRecurringTransactionsArguments{
			AccountId:            bankAccount.AccountId,
			BankAccountId:        bankAccount.BankAccountId,
			TransactionClusterId: cluster.TransactionClusterId,
		}
	}
	return result
}

func TestCalculateTransactionClusters(t *testing.T) {
	t.Run("enqueues recurring calculations for each cluster", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		givenIHaveMonthlyTransactions(t, clock, bankAccount, "GITHUB INC", 800, 6)
		givenIHaveMonthlyTransactions(t, clock, bankAccount, "SENTRY IO", 2600, 6)

		var enqueued []any
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(recurring_jobs.CalculateRecurringTransactions),
				gomock.Any(),
				gomock.Any(),
			).
			Do(func(_ context.Context, _ string, _ time.Time, args []any) {
				enqueued = args
			}).
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

		expected := recurringArgumentsForClusters(t, bankAccount)
		assert.Len(t, expected, 2, "there should be a cluster for each merchant")
		assert.ElementsMatch(t, expected, enqueued, "should enqueue one recurring calculation for each cluster")
	})

	t.Run("does not enqueue clusters that were removed", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
		givenIHaveMonthlyTransactions(t, clock, bankAccount, "GITHUB INC", 800, 6)
		sentry := givenIHaveMonthlyTransactions(t, clock, bankAccount, "SENTRY IO", 2600, 6)

		var firstEnqueued, secondEnqueued []any
		firstCall := enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(recurring_jobs.CalculateRecurringTransactions),
				gomock.Any(),
				gomock.Any(),
			).
			Do(func(_ context.Context, _ string, _ time.Time, args []any) {
				firstEnqueued = args
			}).
			Return(nil).
			Times(1)
		enqueuer.EXPECT().
			BulkEnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(recurring_jobs.CalculateRecurringTransactions),
				gomock.Any(),
				gomock.Any(),
			).
			Do(func(_ context.Context, _ string, _ time.Time, args []any) {
				secondEnqueued = args
			}).
			After(firstCall).
			Return(nil).
			Times(1)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(2)
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
		assert.Len(t, firstEnqueued, 2, "should enqueue a recurring calculation for both clusters")

		// Delete all of the sentry transactions, the next time clusters are
		// calculated the sentry cluster won't exist anymore. Only the deleted at is
		// set so the cluster ID the job assigned is left alone.
		sentryIds := make([]models.ID[models.Transaction], len(sentry))
		for i := range sentry {
			sentryIds[i] = sentry[i].TransactionId
		}
		_, err = db.NewUpdate().
			Model(&models.Transaction{}).
			Set(`"deleted_at" = ?`, clock.Now()).
			Where(`"transaction"."account_id" = ?`, bankAccount.AccountId).
			Where(`"transaction"."transaction_id" IN (?)`, bun.List(sentryIds)).
			Exec(t.Context())
		require.NoError(t, err, "must be able to delete the sentry transactions")

		err = similar_jobs.CalculateTransactionClusters(
			mockqueue.NewMockContext(context),
			similar_jobs.CalculateTransactionClustersArguments{
				AccountId:     bankAccount.AccountId,
				BankAccountId: bankAccount.BankAccountId,
			},
		)
		assert.NoError(t, err, "must be able to recalculate transaction clusters")

		expected := recurringArgumentsForClusters(t, bankAccount)
		assert.Len(t, expected, 1, "only the github cluster should still exist")
		assert.ElementsMatch(t, expected, secondEnqueued, "should only enqueue the cluster that still exists")
	})
}
