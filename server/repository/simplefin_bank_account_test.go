package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_CreateSimpleFinBankAccount(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		bankAccount := models.SimpleFinBankAccount{
			SimpleFinLinkId:  *link.SimpleFinLinkId,
			SimpleFinId:      "ACT-1234",
			ConnectionId:     "CON-1234",
			Name:             "Checking",
			InstitutionName:  "Lehman Brothers",
			Currency:         "USD",
			Status:           models.SimpleFinBankAccountStatusInactive,
			CurrentBalance:   10000,
			AvailableBalance: 9000,
		}
		err := repo.CreateSimpleFinBankAccount(t.Context(), &bankAccount)
		assert.NoError(t, err, "should be able to create a simplefin bank account")
		assert.NotZero(t, bankAccount.SimpleFinBankAccountId, "should have generated a simplefin bank account Id")
		assert.Equal(t, user.AccountId, bankAccount.AccountId, "account Id should be set by the repository")
		assert.Equal(t, user.UserId, bankAccount.CreatedBy, "created by should be set by the repository")
	})
}

func TestRepositoryBase_UpdateSimpleFinBankAccount(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		simpleFinBankAccount := bankAccount.SimpleFinBankAccount
		simpleFinBankAccount.Status = models.SimpleFinBankAccountStatusInactive
		simpleFinBankAccount.CurrentBalance = 12345
		err := repo.UpdateSimpleFinBankAccount(t.Context(), simpleFinBankAccount)
		assert.NoError(t, err, "should be able to update the simplefin bank account")

		result, err := repo.GetSimpleFinBankAccountForSimpleFinLink(
			t.Context(),
			*link.SimpleFinLinkId,
			simpleFinBankAccount.SimpleFinBankAccountId,
		)
		assert.NoError(t, err, "should be able to read the simplefin bank account back")
		assert.Equal(t, models.SimpleFinBankAccountStatusInactive, result.Status, "status should be updated")
		assert.EqualValues(t, 12345, result.CurrentBalance, "balance should be updated")
	})

	t.Run("cant update someone elses bank account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			db,
			log,
		)

		simpleFinBankAccount := *bankAccount.SimpleFinBankAccount
		simpleFinBankAccount.CurrentBalance = 12345
		err := repo.UpdateSimpleFinBankAccount(t.Context(), &simpleFinBankAccount)
		assert.EqualError(t, err, "no rows updated", "should not update another account's bank account")
	})
}

func TestRepositoryBase_GetSimpleFinBankAccountForSimpleFinLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		result, err := repo.GetSimpleFinBankAccountForSimpleFinLink(
			t.Context(),
			*link.SimpleFinLinkId,
			*bankAccount.SimpleFinBankAccountId,
		)
		assert.NoError(t, err, "should be able to retrieve the simplefin bank account")
		assert.Equal(t, *bankAccount.SimpleFinBankAccountId, result.SimpleFinBankAccountId, "should return the right bank account")
	})

	t.Run("wrong link", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		otherLink := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		result, err := repo.GetSimpleFinBankAccountForSimpleFinLink(
			t.Context(),
			*otherLink.SimpleFinLinkId,
			*bankAccount.SimpleFinBankAccountId,
		)
		assert.EqualError(t, err, "failed to retrieve SimpleFIN bank account: sql: no rows in result set", "should not find a bank account on another link")
		assert.Nil(t, result, "should not return a bank account")
	})

	t.Run("deleted", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		bankAccount.SimpleFinBankAccount.Status = models.SimpleFinBankAccountStatusInactive
		bankAccount.SimpleFinBankAccount.DeletedAt = new(clock.Now())
		err := repo.UpdateSimpleFinBankAccount(t.Context(), bankAccount.SimpleFinBankAccount)
		assert.NoError(t, err, "must be able to soft delete the simplefin bank account")

		result, err := repo.GetSimpleFinBankAccountForSimpleFinLink(
			t.Context(),
			*link.SimpleFinLinkId,
			*bankAccount.SimpleFinBankAccountId,
		)
		assert.EqualError(t, err, "failed to retrieve SimpleFIN bank account: sql: no rows in result set", "should not find a deleted bank account")
		assert.Nil(t, result, "should not return a bank account")
	})

	t.Run("get bank account includes simplefin bank account", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		result, err := repo.GetBankAccount(t.Context(), bankAccount.BankAccountId)
		assert.NoError(t, err, "should be able to retrieve the bank account")
		if assert.NotNil(t, result.SimpleFinBankAccount, "simplefin bank account should be loaded") {
			assert.Equal(t, *bankAccount.SimpleFinBankAccountId, result.SimpleFinBankAccount.SimpleFinBankAccountId, "should load the right simplefin bank account")
		}
	})
}

func TestRepositoryBase_GetSimpleFinBankAccountsBySimpleFinLink(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		result, err := repo.GetSimpleFinBankAccountsBySimpleFinLink(t.Context(), *link.SimpleFinLinkId)
		assert.NoError(t, err, "should be able to list simplefin bank accounts")
		assert.Len(t, result, 2, "should return both bank accounts")
	})

	t.Run("excludes deleted", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		bankAccount.SimpleFinBankAccount.Status = models.SimpleFinBankAccountStatusInactive
		bankAccount.SimpleFinBankAccount.DeletedAt = new(clock.Now())
		err := repo.UpdateSimpleFinBankAccount(t.Context(), bankAccount.SimpleFinBankAccount)
		assert.NoError(t, err, "must be able to soft delete the simplefin bank account")

		result, err := repo.GetSimpleFinBankAccountsBySimpleFinLink(t.Context(), *link.SimpleFinLinkId)
		assert.NoError(t, err, "should be able to list simplefin bank accounts")
		assert.Len(t, result, 1, "deleted bank account should not be returned")
		assert.NotEqual(t, *bankAccount.SimpleFinBankAccountId, result[0].SimpleFinBankAccountId, "should only return the bank account that wasnt deleted")
	})

	t.Run("no cross account reads", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		otherUser, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewRepositoryFromSession(
			clock,
			otherUser.UserId,
			otherUser.AccountId,
			db,
			log,
		)

		result, err := repo.GetSimpleFinBankAccountsBySimpleFinLink(t.Context(), *link.SimpleFinLinkId)
		assert.NoError(t, err, "should be able to list simplefin bank accounts")
		assert.Empty(t, result, "should not see another account's bank accounts")
	})
}

func TestRepositoryBase_GetAllSimpleFinBankAccountsBySimpleFinLink(t *testing.T) {
	t.Run("includes deleted", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveASimpleFinLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)
		fixtures.GivenIHaveASimpleFinBankAccount(t, clock, &link)

		repo := repository.NewRepositoryFromSession(
			clock,
			user.UserId,
			user.AccountId,
			db,
			log,
		)

		bankAccount.SimpleFinBankAccount.Status = models.SimpleFinBankAccountStatusInactive
		bankAccount.SimpleFinBankAccount.DeletedAt = new(clock.Now())
		err := repo.UpdateSimpleFinBankAccount(t.Context(), bankAccount.SimpleFinBankAccount)
		assert.NoError(t, err, "must be able to soft delete the simplefin bank account")

		result, err := repo.GetAllSimpleFinBankAccountsBySimpleFinLink(t.Context(), *link.SimpleFinLinkId)
		assert.NoError(t, err, "should be able to list simplefin bank accounts")
		assert.Len(t, result, 2, "deleted bank account should still be returned")
	})
}
