package repository_test

import (
	"slices"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryBase_GetTransactionClusterIds(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		otherBankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.SavingsBankAccountSubType,
		)
		expected := []models.ID[models.TransactionCluster]{
			givenIHaveATransactionCluster(t, bankAccount).TransactionClusterId,
			givenIHaveATransactionCluster(t, bankAccount).TransactionClusterId,
			givenIHaveATransactionCluster(t, bankAccount).TransactionClusterId,
		}
		slices.Sort(expected)
		givenIHaveATransactionCluster(t, otherBankAccount)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		result, err := repo.GetTransactionClusterIds(t.Context(), bankAccount.BankAccountId)
		assert.NoError(t, err, "must be able to read cluster ids")
		assert.Equal(t, expected, result, "should be the bank account's clusters ordered by id")
	})

	t.Run("no clusters", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		result, err := repo.GetTransactionClusterIds(t.Context(), bankAccount.BankAccountId)
		assert.NoError(t, err, "must be able to read cluster ids")
		assert.NotNil(t, result, "should be an empty slice, not nil")
		assert.Empty(t, result, "there should be no clusters")
	})

	t.Run("cant read other accounts", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)
		givenIHaveATransactionCluster(t, bankAccount)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			testutils.GetPgDatabase(t),
			log,
		)

		result, err := repo.GetTransactionClusterIds(t.Context(), bankAccount.BankAccountId)
		assert.NoError(t, err, "must be able to read cluster ids")
		assert.Empty(t, result, "should not see another account's clusters")
	})
}

func TestRepositoryBase_GetTransactionClusters(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
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
		assert.NoError(t, err, "must retrieve transaction clusters")
		require.Len(t, clusters, 3, "should return every cluster in the bank account")
		assert.Equal(t, "Wendy's", clusters[0].Name, "should be ordered by name descending")
		assert.Equal(t, "Starbucks", clusters[1].Name, "should be ordered by name descending")
		assert.Equal(t, "Amazon", clusters[2].Name, "should be ordered by name descending")
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
		require.Len(t, all, 5, "must have every cluster to compare against")

		firstPage, err := repo.GetTransactionClusters(t.Context(), bank.BankAccountId, 3, 0)
		assert.NoError(t, err, "must retrieve first page")
		assert.Len(t, firstPage, 3, "first page should be full")

		secondPage, err := repo.GetTransactionClusters(t.Context(), bank.BankAccountId, 3, 3)
		assert.NoError(t, err, "must retrieve second page")
		assert.Len(t, secondPage, 2, "second page should have whats left")

		// Pages should line up with the unpaginated result
		assert.Equal(t, all[:3], firstPage, "first page should match")
		assert.Equal(t, all[3:], secondPage, "second page should match")
	})

	t.Run("scoped to bank account", func(t *testing.T) {
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
		assert.NoError(t, err, "must retrieve transaction clusters")
		require.Len(t, clusters, 1, "should only return the checking account cluster")
		assert.Equal(t, checkingCluster.TransactionClusterId, clusters[0].TransactionClusterId, "should be the checking cluster")
	})

	t.Run("scoped to account", func(t *testing.T) {
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
		assert.NoError(t, err, "must not fail when the bank account belongs to someone else")
		assert.Empty(t, clusters, "should not return clusters from another account")
	})
}
