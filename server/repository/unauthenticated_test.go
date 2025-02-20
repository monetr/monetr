package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/consts"
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
		Locale:                  "en_US",
		StripeCustomerId:        nil,
		StripeSubscriptionId:    nil,
		SubscriptionActiveUntil: nil,
	}
	err := repo.CreateAccountV2(context.Background(), &account)
	assert.NoError(t, err, "should successfully create account")
	assert.NotEmpty(t, account, "new account should not be empty")
	assert.NotEmpty(t, account.AccountId, "account Id should have been generated")
}

func TestUnauthenticatedRepo_CreateLogin(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
		login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.NotEmpty(t, login.LoginId, "login Id should have been generated")
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
		assert.NotEmpty(t, loginOne.LoginId, "login Id should have been generated")

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
		assert.NotEmpty(t, login.LoginId, "login Id should have been generated")

		account := models.Account{
			Timezone:                time.UTC.String(),
			Locale:                  "en_US",
			StripeCustomerId:        nil,
			StripeSubscriptionId:    nil,
			SubscriptionActiveUntil: nil,
		}
		err = repo.CreateAccountV2(context.Background(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.NotEmpty(t, account.AccountId, "account Id should have been generated")

		user := models.User{
			LoginId:   login.LoginId,
			AccountId: account.AccountId,
			Role:      models.UserRoleOwner,
		}
		err = repo.CreateUser(context.Background(), &user)
		assert.NoError(t, err, "should successfully create user")
		assert.NotEmpty(t, user, "new user should not be empty")
		assert.NotEmpty(t, user.UserId, "user Id should have been generated")
	})

	t.Run("unique login per account", func(t *testing.T) {
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)
		email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)

		login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
		assert.NoError(t, err, "should successfully create login")
		assert.NotEmpty(t, login, "new login should not be empty")
		assert.NotEmpty(t, login.LoginId, "login Id should have been generated")

		account := models.Account{
			Timezone:                time.UTC.String(),
			Locale:                  "en_US",
			StripeCustomerId:        nil,
			StripeSubscriptionId:    nil,
			SubscriptionActiveUntil: nil,
		}
		err = repo.CreateAccountV2(context.Background(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.NotEmpty(t, account.AccountId, "account Id should have been generated")

		user := models.User{
			LoginId:   login.LoginId,
			AccountId: account.AccountId,
			Role:      models.UserRoleOwner,
		}
		err = repo.CreateUser(context.Background(), &user)
		assert.NoError(t, err, "should successfully create user")
		assert.NotEmpty(t, user, "new user should not be empty")
		assert.NotEmpty(t, user.UserId, "user Id should have been generated")

		// Try to create another user with the same login and account, this should fail.
		userAgain := user
		userAgain.UserId = ""
		err = repo.CreateUser(context.Background(), &userAgain)
		assert.Error(t, err, "should not create duplicate login for account")
	})

	t.Run("unique owner per account", func(t *testing.T) {
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)

		account := models.Account{
			Timezone:                time.UTC.String(),
			Locale:                  "en_US",
			StripeCustomerId:        nil,
			StripeSubscriptionId:    nil,
			SubscriptionActiveUntil: nil,
		}
		err := repo.CreateAccountV2(context.Background(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.NotEmpty(t, account.AccountId, "account Id should have been generated")

		{ // Create the first user, and make it the owner.
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")

			user := models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				Role:      models.UserRoleOwner,
			}
			err = repo.CreateUser(context.Background(), &user)
			assert.NoError(t, err, "should successfully create user")
			assert.NotEmpty(t, user, "new user should not be empty")
			assert.NotEmpty(t, user.UserId, "user Id should have been generated")
		}

		{ // Now create a different user and try to make it the owner as well.
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")

			user := models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				Role:      models.UserRoleOwner,
			}
			err = repo.CreateUser(context.Background(), &user)
			assert.Error(t, err, "cannot create another user who is also an owner of the same account")

			user.Role = models.UserRoleMember
			err = repo.CreateUser(context.Background(), &user)
			assert.NoError(t, err, "but can create the user as a member of the account")
		}
	})

	t.Run("missing role", func(t *testing.T) {
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)

		account := models.Account{
			Timezone:                time.UTC.String(),
			Locale:                  "en_US",
			StripeCustomerId:        nil,
			StripeSubscriptionId:    nil,
			SubscriptionActiveUntil: nil,
		}
		err := repo.CreateAccountV2(context.Background(), &account)
		assert.NoError(t, err, "should successfully create account")
		assert.NotEmpty(t, account, "new account should not be empty")
		assert.NotEmpty(t, account.AccountId, "account Id should have been generated")

		{ // Create the first user, and make it the owner.
			email, password := gofakeit.Email(), gofakeit.Password(true, true, true, true, false, 32)
			login, err := repo.CreateLogin(context.Background(), email, password, gofakeit.FirstName(), gofakeit.LastName())
			assert.NoError(t, err, "should successfully create login")
			assert.NotEmpty(t, login, "new login should not be empty")
			assert.NotEmpty(t, login.LoginId, "login Id should have been generated")

			user := models.User{
				LoginId:   login.LoginId,
				AccountId: account.AccountId,
				// Don't include a role
			}
			err = repo.CreateUser(context.Background(), &user)
			assert.Error(t, err, "should not be able to create a user without a role")
		}
	})
}

func TestUnauthenticatedRepo_ResetPassword(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)
		login, _ := fixtures.GivenIHaveLogin(t, clock)

		err := repo.ResetPassword(context.Background(), login.LoginId, gofakeit.UUID())
		assert.NoError(t, err, "must reset password without an error")
	})

	t.Run("bad login", func(t *testing.T) {
		clock := clock.NewMock()
		repo := GetTestUnauthenticatedRepository(t, clock)

		err := repo.ResetPassword(context.Background(), "lgn_bogus", gofakeit.UUID())
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
