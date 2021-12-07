package repository_test

import (
	"context"
	"testing"

	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryBase_GetAccount(t *testing.T) {
	t.Run("account does not exist", func(t *testing.T) {
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		repo := repository.NewRepositoryFromSession(123, 1234, db)
		account, err := repo.GetAccount(context.Background())
		assert.EqualError(t, err, "failed to retrieve account: pg: no rows in result set")
		assert.Nil(t, account, "should not return an account")
	})

	t.Run("account exists", func(t *testing.T) {
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t)

		repo := repository.NewRepositoryFromSession(user.UserId, user.AccountId, db)
		account, err := repo.GetAccount(context.Background())
		assert.NoError(t, err, "must not return an error if the account exists")
		assert.NotNil(t, account, "should return a valid account")
		assert.Equal(t, user.AccountId, account.AccountId, "retrieved account must match the fixture")
	})

	t.Run("will cache the account", func(t *testing.T) {
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t)

		repo := repository.NewRepositoryFromSession(user.UserId, user.AccountId, db)
		account, err := repo.GetAccount(context.Background())
		assert.NoError(t, err, "must not return an error if the account exists")
		assert.NotNil(t, account, "should return a valid account")
		assert.Equal(t, user.AccountId, account.AccountId, "retrieved account must match the fixture")

		// Then delete the account from the database.
		{
			result, err := db.ModelContext(context.Background(), &models.User{
				UserId:    user.UserId,
				AccountId: user.AccountId,
			}).
				WherePK().
				Delete()
			assert.NoError(t, err, "must successfully delete the account to test the cache")
			assert.Equal(t, 1, result.RowsAffected(), "should only delete the one account")

			result, err = db.ModelContext(context.Background(), &models.Account{AccountId: user.AccountId}).
				WherePK().
				Delete()
			assert.NoError(t, err, "must successfully delete the account to test the cache")
			assert.Equal(t, 1, result.RowsAffected(), "should only delete the one account")
		}

		// Calling get account again should still return the account as it is cached.
		account, err = repo.GetAccount(context.Background())
		assert.NoError(t, err, "must not return an error if the account exists")
		assert.NotNil(t, account, "should return a valid account")
		assert.Equal(t, user.AccountId, account.AccountId, "retrieved account must match the fixture")
	})
}
