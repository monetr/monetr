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
)

func TestRepositoryBase_GetTransactionClusterIds(t *testing.T) {
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

	t.Run("only the specified bank account, in order", func(t *testing.T) {
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

	t.Run("cannot read another account's clusters", func(t *testing.T) {
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
