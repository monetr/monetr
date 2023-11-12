package repository_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/hash"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func GetTestUnauthenticatedRepository(t *testing.T, clock clock.Clock) repository.UnauthenticatedRepository {
	db := testutils.GetPgDatabase(t)
	return repository.NewUnauthenticatedRepository(clock, db)
}

func TestUnauthenticatedRepo_CreateAccount(t *testing.T) {
	clock := clock.NewMock()
	repo := GetTestUnauthenticatedRepository(t, clock)
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
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
		login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.Greater(t, login.LoginId, uint64(0), "loginId should be greater than 0")
	})

	t.Run("duplicate email", func(t *testing.T) {
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)
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
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)
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
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)
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
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)
		login, _ := fixtures.GivenIHaveLogin(t, clock)

		err := repo.ResetPassword(context.Background(), login.LoginId, hash.HashPassword(login.Email, "new Password"))
		assert.NoError(t, err, "must reset password without an error")
	})

	t.Run("bad login", func(t *testing.T) {
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)

		err := repo.ResetPassword(context.Background(), math.MaxUint64, hash.HashPassword(testutils.GetUniqueEmail(t), "new Password"))
		assert.EqualError(t, err, "no logins were updated", "should return an error for invalid login")
	})
}

func seedLogin(t *testing.T, login *models.Login) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(gofakeit.Password(true, true, true, true, false, 16)),
		consts.BcryptCost,
	)
	require.NoError(t, err, "must not have an error when generating the password")
	loginWithPassword := models.LoginWithHash{
		Login: *login,
		Crypt: hashedPassword,
	}

	db := testutils.GetPgDatabase(t)
	result, err := db.Model(&loginWithPassword).Insert(&loginWithPassword)
	assert.NoError(t, err, "must insert login for test")
	assert.Equal(t, 1, result.RowsAffected(), "must affect 1 row for the insert")

	login.LoginId = loginWithPassword.LoginId
}

func TestEmailRepositoryBase_SetEmailVerified(t *testing.T) {
	assertEmailVerified := func(t *testing.T, emailAddress string, verified bool) {
		db := testutils.GetPgDatabase(t)
		exists, err := db.Model(&models.Login{}).
			Where(`"login"."email" = ?`, emailAddress).
			Where(`"login"."is_email_verified" = ?`, verified).
			Limit(1).
			Exists()
		assert.NoError(t, err, "must assert that the email is verified")
		assert.True(t, exists, "login must be in the expected state")
	}

	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()

		emailAddress := testutils.GetUniqueEmail(t)

		seedLogin(t, &models.Login{
			Email:           emailAddress,
			FirstName:       gofakeit.FirstName(),
			LastName:        gofakeit.LastName(),
			IsEnabled:       true,
			IsEmailVerified: false,
		})

		repo := GetTestUnauthenticatedRepository(t, clock)

		assertEmailVerified(t, emailAddress, repository.EmailNotVerified)

		err := repo.SetEmailVerified(context.Background(), emailAddress)
		assert.NoError(t, err, "email must successfully be verified")

		assertEmailVerified(t, emailAddress, repository.EmailVerified)
	})

	t.Run("login with email does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		emailAddress := testutils.GetUniqueEmail(t)

		repo := GetTestUnauthenticatedRepository(t, clock)

		err := repo.SetEmailVerified(context.Background(), emailAddress)
		assert.EqualError(t, err, "email cannot be verified")
	})

	t.Run("email already verified", func(t *testing.T) {
		clock := clock.NewMock()
		emailAddress := testutils.GetUniqueEmail(t)

		seedLogin(t, &models.Login{
			Email:           emailAddress,
			FirstName:       gofakeit.FirstName(),
			LastName:        gofakeit.LastName(),
			IsEnabled:       true,
			IsEmailVerified: true,
		})

		repo := GetTestUnauthenticatedRepository(t, clock)

		assertEmailVerified(t, emailAddress, repository.EmailVerified)

		err := repo.SetEmailVerified(context.Background(), emailAddress)
		assert.EqualError(t, err, "email cannot be verified")

		assertEmailVerified(t, emailAddress, repository.EmailVerified)
	})

	t.Run("login not enabled", func(t *testing.T) {
		clock := clock.NewMock()
		emailAddress := testutils.GetUniqueEmail(t)

		seedLogin(t, &models.Login{
			Email:           emailAddress,
			FirstName:       gofakeit.FirstName(),
			LastName:        gofakeit.LastName(),
			IsEnabled:       false,
			IsEmailVerified: false,
		})

		repo := GetTestUnauthenticatedRepository(t, clock)

		assertEmailVerified(t, emailAddress, repository.EmailNotVerified)

		err := repo.SetEmailVerified(context.Background(), emailAddress)
		assert.EqualError(t, err, "email cannot be verified")

		assertEmailVerified(t, emailAddress, repository.EmailNotVerified)
	})
}

func TestEmailRepositoryBase_GetLoginForEmail(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		emailAddress := testutils.GetUniqueEmail(t)

		originalLogin := &models.Login{
			Email:     emailAddress,
			FirstName: gofakeit.FirstName(),
			LastName:  gofakeit.LastName(),
		}
		seedLogin(t, originalLogin)

		repo := GetTestUnauthenticatedRepository(t, clock)

		login, err := repo.GetLoginForEmail(context.Background(), emailAddress)
		assert.NoError(t, err, "must retrieve login for email successfully")
		assert.NotNil(t, login, "login result should not be nil")
		assert.Equal(t, originalLogin.LoginId, login.LoginId, "login Id should match expected")
	})

	t.Run("email does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		emailAddress := testutils.GetUniqueEmail(t)

		repo := GetTestUnauthenticatedRepository(t, clock)

		login, err := repo.GetLoginForEmail(context.Background(), emailAddress)
		assert.EqualError(t, err, "failed to retrieve login by email: pg: no rows in result set")
		assert.Nil(t, login, "login result should be nil")
	})
}
