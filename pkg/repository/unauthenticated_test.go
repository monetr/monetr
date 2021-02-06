package repository

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func GetTestUnauthenticatedRepository(t *testing.T) UnauthenticatedRepository {
	txn := testutils.GetPgDatabaseTxn(t)
	return NewUnauthenticatedRepository(txn)
}

func TestUnauthenticatedRepo_CreateAccount(t *testing.T) {
	repo := GetTestUnauthenticatedRepository(t)
	account, err := repo.CreateAccount(time.UTC)
	assert.NoError(t, err, "should successfully create account")
	assert.NotEmpty(t, account, "new account should not be empty")
	assert.Greater(t, 0, account.AccountId, "accountId should be greater than 0")
}

func TestUnauthenticatedRepo_CreateLogin(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
		hash := testutils.MustHashLogin(t, email, password)
		login, err := repo.CreateLogin(email, hash, true)
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.Greater(t, login.LoginId, uint64(0), "loginId should be greater than 0")
	})

	t.Run("duplicate email", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		email := gofakeit.Email()

		passwordOne := gofakeit.Password(true, true, true, true, false, 32)
		hashOne := testutils.MustHashLogin(t, email, passwordOne)

		// Creating the first login should succeed.
		loginOne, err := repo.CreateLogin(email, hashOne, true)
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, loginOne, "new login should not be empty")
		assert.Greater(t, loginOne.LoginId, uint64(0), "loginId should be greater than 0")

		passwordTwo := gofakeit.Password(true, true, true, true, false, 32)
		hashTwo := testutils.MustHashLogin(t, email, passwordTwo)

		// Creating the first login should succeed.
		loginTwo, err := repo.CreateLogin(email, hashTwo, true)
		assert.Error(t, err, "should fail to create another login with the same email")
		assert.EqualError(t, err, "a login with the same email already exists")
		assert.Nil(t, loginTwo, "should return nil for login")
	})
}

func TestUnauthenticatedRepo_CreateUser(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
		hash := testutils.MustHashLogin(t, email, password)

		login, err := repo.CreateLogin(email, hash, true)
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.Greater(t, login.LoginId, uint64(0), "loginId should be greater than 0")

		account, err := repo.CreateAccount(time.UTC)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.Greater(t, account.AccountId, uint64(0), "accountId should be greater than 0")

		firstName, lastName := gofakeit.FirstName(), gofakeit.LastName()
		user, err := repo.CreateUser(login.LoginId, account.AccountId, firstName, lastName)
		assert.NoError(t, err, "should successfully create user")
		assert.NotEmpty(t, user, "new user should not be empty")
		assert.Greater(t, user.UserId, uint64(0), "userId should be greater than 0")
	})

	t.Run("unique login per account", func(t *testing.T) {
		repo := GetTestUnauthenticatedRepository(t)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
		hash := testutils.MustHashLogin(t, email, password)

		login, err := repo.CreateLogin(email, hash, true)
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.Greater(t, login.LoginId, uint64(0), "loginId should be greater than 0")

		account, err := repo.CreateAccount(time.UTC)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.Greater(t, account.AccountId, uint64(0), "accountId should be greater than 0")

		firstName, lastName := gofakeit.FirstName(), gofakeit.LastName()
		user, err := repo.CreateUser(login.LoginId, account.AccountId, firstName, lastName)
		assert.NoError(t, err, "should successfully create user")
		assert.NotEmpty(t, user, "new user should not be empty")
		assert.Greater(t, user.UserId, uint64(0), "userId should be greater than 0")

		// Try to create another user with the same login and account, this should fail.
		userAgain, err := repo.CreateUser(login.LoginId, account.AccountId, firstName, lastName)
		assert.Error(t, err, "should not create duplicate login for account")
		assert.Nil(t, userAgain, "should return nil for user")
	})
}
