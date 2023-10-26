package repository_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/hash"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
)

func GetTestUnauthenticatedRepository(t *testing.T) repository.UnauthenticatedRepository {
	txn := testutils.GetPgDatabaseTxn(t)
	return repository.NewUnauthenticatedRepository(txn)
}

func TestUnauthenticatedRepo_CreateAccount(t *testing.T) {
	repo := GetTestUnauthenticatedRepository(t)
	account := models.Account{
		Timezone:                time.UTC.String(),
		StripeCustomerId:        nil,
		StripeSubscriptionId:    nil,
		SubscriptionActiveUntil: nil,
	}
	err := repo.CreateAccountV2(context.Background(), &account)
	assert.NoError(t, err, "should successfully create account")
	assert.NotEmpty(t, account, "new account should not be empty")
	assert.Greater(t, account.AccountId, uint64(0), "accountId should be greater than 0")
}

func TestUnauthenticatedRepo_CreateLogin(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
		login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.Greater(t, login.LoginId, uint64(0), "loginId should be greater than 0")
	})

	t.Run("duplicate email", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		email := gofakeit.Email()

		passwordOne := gofakeit.Password(true, true, true, true, false, 32)

		// Creating the first login should succeed.
		loginOne, err := repo.CreateLogin(context.Background(), email, passwordOne, gofakeit.FirstName(), gofakeit.LastName())
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, loginOne, "new login should not be empty")
		assert.Greater(t, loginOne.LoginId, uint64(0), "loginId should be greater than 0")

		passwordTwo := gofakeit.Password(true, true, true, true, false, 32)

		// Creating the first login should succeed.
		loginTwo, err := repo.CreateLogin(context.Background(), email, passwordTwo, gofakeit.FirstName(), gofakeit.LastName())
		assert.Error(t, err, "should fail to create another login with the same email")
		assert.EqualError(t, err, "a login with the same email already exists")
		assert.Nil(t, loginTwo, "should return nil for login")
	})
}

func TestUnauthenticatedRepo_CreateUser(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)

		login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.Greater(t, login.LoginId, uint64(0), "loginId should be greater than 0")

		account := models.Account{
			Timezone:                time.UTC.String(),
			StripeCustomerId:        nil,
			StripeSubscriptionId:    nil,
			SubscriptionActiveUntil: nil,
		}
		err = repo.CreateAccountV2(context.Background(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.Greater(t, account.AccountId, uint64(0), "accountId should be greater than 0")

		firstName, lastName := gofakeit.FirstName(), gofakeit.LastName()
		user := models.User{
			LoginId:          login.LoginId,
			AccountId:        account.AccountId,
			FirstName:        firstName,
			LastName:         lastName,
			StripeCustomerId: nil,
		}
		err = repo.CreateUser(context.Background(), login.LoginId, account.AccountId, &user)
		assert.NoError(t, err, "should successfully create user")
		assert.NotEmpty(t, user, "new user should not be empty")
		assert.Greater(t, user.UserId, uint64(0), "userId should be greater than 0")
	})

	t.Run("unique login per account", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)

		login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.Greater(t, login.LoginId, uint64(0), "loginId should be greater than 0")

		account := models.Account{
			Timezone:                time.UTC.String(),
			StripeCustomerId:        nil,
			StripeSubscriptionId:    nil,
			SubscriptionActiveUntil: nil,
		}
		err = repo.CreateAccountV2(context.Background(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.Greater(t, account.AccountId, uint64(0), "accountId should be greater than 0")

		firstName, lastName := gofakeit.FirstName(), gofakeit.LastName()
		user := models.User{
			LoginId:          login.LoginId,
			AccountId:        account.AccountId,
			FirstName:        firstName,
			LastName:         lastName,
			StripeCustomerId: nil,
		}
		err = repo.CreateUser(context.Background(), login.LoginId, account.AccountId, &user)
		assert.NoError(t, err, "should successfully create user")
		assert.NotEmpty(t, user, "new user should not be empty")
		assert.Greater(t, user.UserId, uint64(0), "userId should be greater than 0")

		// Try to create another user with the same login and account, this should fail.
		userAgain := user
		userAgain.UserId = 0
		err = repo.CreateUser(context.Background(), login.LoginId, account.AccountId, &userAgain)
		assert.Error(t, err, "should not create duplicate login for account")
		assert.Zero(t, userAgain.UserId, "should not have an id")
	})
}

func TestUnauthenticatedRepo_ResetPassword(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		login, _ := fixtures.GivenIHaveLogin(t)

		err := repo.ResetPassword(context.Background(), login.LoginId, hash.HashPassword(login.Email, "new Password"))
		assert.NoError(t, err, "must reset password without an error")
	})

	t.Run("bad login", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)

		err := repo.ResetPassword(context.Background(), math.MaxUint64, hash.HashPassword(testutils.GetUniqueEmail(t), "new Password"))
		assert.EqualError(t, err, "no logins were updated", "should return an error for invalid login")
	})
}
