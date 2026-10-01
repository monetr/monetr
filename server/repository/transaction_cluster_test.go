package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryBase_GetTransactionClusters(t *testing.T) {
	t.Run("ordered by name descending", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bank := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		transactions := fixtures.GivenIHaveNTransactions(t, clock, bank, 3)

		names := []string{"Starbucks", "Amazon", "Wendy's"}
		for i, name := range names {
			testutils.MustInsert(t, models.TransactionCluster{
				AccountId:     bank.AccountId,
				BankAccountId: bank.BankAccountId,
				Name:          name,
				OriginalName:  name,
				Members:       []models.ID[models.Transaction]{transactions[i].TransactionId},
			})
		}

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		clusters, err := repo.GetTransactionClusters(t.Context(), bank.BankAccountId, 10, 0)
		require.NoError(t, err, "must retrieve transaction clusters")
		require.Len(t, clusters, 3, "should return every cluster in the bank account")
		assert.Equal(t, "Wendy's", clusters[0].Name)
		assert.Equal(t, "Starbucks", clusters[1].Name)
		assert.Equal(t, "Amazon", clusters[2].Name)
	})

	t.Run("limit and offset", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bank := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		for _, transaction := range fixtures.GivenIHaveNTransactions(t, clock, bank, 5) {
			testutils.MustInsert(t, models.TransactionCluster{
				AccountId:     bank.AccountId,
				BankAccountId: bank.BankAccountId,
				Name:          transaction.Name,
				OriginalName:  transaction.OriginalName,
				Members:       []models.ID[models.Transaction]{transaction.TransactionId},
			})
		}

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		all, err := repo.GetTransactionClusters(t.Context(), bank.BankAccountId, 10, 0)
		require.NoError(t, err, "must retrieve all transaction clusters")
		require.Len(t, all, 5)

		firstPage, err := repo.GetTransactionClusters(t.Context(), bank.BankAccountId, 3, 0)
		require.NoError(t, err, "must retrieve first page")
		require.Len(t, firstPage, 3)

		secondPage, err := repo.GetTransactionClusters(t.Context(), bank.BankAccountId, 3, 3)
		require.NoError(t, err, "must retrieve second page")
		require.Len(t, secondPage, 2)

		// Pages should line up with the unpaginated result.
		assert.Equal(t, all[:3], firstPage)
		assert.Equal(t, all[3:], secondPage)
	})

	t.Run("scoped to the bank account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		checking := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		savings := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.SavingsBankAccountSubType,
		)

		checkingTransaction := fixtures.GivenIHaveATransaction(t, clock, checking)
		checkingCluster := testutils.MustInsert(t, models.TransactionCluster{
			AccountId:     checking.AccountId,
			BankAccountId: checking.BankAccountId,
			Name:          checkingTransaction.Name,
			OriginalName:  checkingTransaction.OriginalName,
			Members:       []models.ID[models.Transaction]{checkingTransaction.TransactionId},
		})

		savingsTransaction := fixtures.GivenIHaveATransaction(t, clock, savings)
		testutils.MustInsert(t, models.TransactionCluster{
			AccountId:     savings.AccountId,
			BankAccountId: savings.BankAccountId,
			Name:          savingsTransaction.Name,
			OriginalName:  savingsTransaction.OriginalName,
			Members:       []models.ID[models.Transaction]{savingsTransaction.TransactionId},
		})

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		clusters, err := repo.GetTransactionClusters(t.Context(), checking.BankAccountId, 10, 0)
		require.NoError(t, err, "must retrieve transaction clusters")
		require.Len(t, clusters, 1, "should only return the checking account cluster")
		assert.Equal(t, checkingCluster.TransactionClusterId, clusters[0].TransactionClusterId)
	})

	t.Run("scoped to the account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		var bank models.BankAccount
		{ // Create a bank account with a cluster under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
			link := fixtures.GivenIHaveAManualLink(t, clock, user)
			bank = fixtures.GivenIHaveABankAccount(
				t,
				clock,
				&link,
				models.DepositoryBankAccountType,
				models.CheckingBankAccountSubType,
			)
			transaction := fixtures.GivenIHaveATransaction(t, clock, bank)
			testutils.MustInsert(t, models.TransactionCluster{
				AccountId:     bank.AccountId,
				BankAccountId: bank.BankAccountId,
				Name:          transaction.Name,
				OriginalName:  transaction.OriginalName,
				Members:       []models.ID[models.Transaction]{transaction.TransactionId},
			})
		}

		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			db,
			log,
		)

		clusters, err := repo.GetTransactionClusters(t.Context(), bank.BankAccountId, 10, 0)
		require.NoError(t, err, "must not fail when the bank account belongs to someone else")
		assert.Empty(t, clusters, "should not return clusters from another account")
	})
}
