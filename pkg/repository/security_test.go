package repository_test

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/hash"
	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBaseSecurityRepository_Login(t *testing.T) {
	t.Run("valid credentials", func(t *testing.T) {
		login, password := fixtures.GivenIHaveLogin(t)
		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t))
		hashedPassword := hash.HashPassword(login.Email, password)

		result, err := repo.Login(context.Background(), login.Email, hashedPassword)
		assert.NoError(t, err, "must not return an error for valid credentials")
		assert.NotNil(t, result, "must return a login object for valid credentials")
		assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
	})

	t.Run("oddly cased email", func(t *testing.T) {
		login, password := fixtures.GivenIHaveLogin(t)
		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t))

		email := strings.ToUpper(login.Email)

		hashedPassword := hash.HashPassword(email, password)

		result, err := repo.Login(context.Background(), email, hashedPassword)
		assert.NoError(t, err, "must not return an error for valid credentials")
		assert.NotNil(t, result, "must return a login object for valid credentials")
		assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
	})

	t.Run("invalid credentials", func(t *testing.T) {
		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t))
		email := testutils.GetUniqueEmail(t)
		password := gofakeit.Generate("????????")
		hashedPassword := hash.HashPassword(email, password)

		result, err := repo.Login(context.Background(), email, hashedPassword)
		assert.EqualError(t, err, "invalid credentials provided")
		assert.Equal(t, repository.ErrInvalidCredentials, errors.Cause(err), "must be caused by invalid credentials")
		assert.Nil(t, result, "must not return a login object when the credentials are invalid")
	})

	t.Run("bad database connection", func(t *testing.T) {
		login, password := fixtures.GivenIHaveLogin(t)
		repo := repository.NewSecurityRepository(testutils.GetBadPgDatabase(t))
		hashedPassword := hash.HashPassword(login.Email, password)

		result, err := repo.Login(context.Background(), login.Email, hashedPassword)
		assert.EqualError(t, err, "failed to verify credentials: forcing a bad connection")
		assert.Nil(t, result, "must not return a result if the connection is bad")
	})
}

func TestBaseSecurityRepository_ChangePassword(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		login, password := fixtures.GivenIHaveLogin(t)
		hashedPassword := hash.HashPassword(login.Email, password)
		newHashedPassword := hash.HashPassword(login.Email, gofakeit.Generate("?????????????"))

		assert.NotEqual(t, hashedPassword, newHashedPassword, "hashed passwords must be different")

		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t))

		{ // Make sure that we can authenticate with the initial hashed password.
			result, err := repo.Login(context.Background(), login.Email, hashedPassword)
			assert.NoError(t, err, "must not return an error for valid credentials")
			assert.NotNil(t, result, "must return a login object for valid credentials")
			assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
		}

		{ // Update the login's password.
			err := repo.ChangePassword(context.Background(), login.LoginId, hashedPassword, newHashedPassword)
			assert.NoError(t, err, "must not return an error when changing the password")
		}

		{ // Make sure that we can no longer authenticate using the old credentials.
			result, err := repo.Login(context.Background(), login.Email, hashedPassword)
			assert.EqualError(t, err, "invalid credentials provided")
			assert.Equal(t, repository.ErrInvalidCredentials, errors.Cause(err), "must be caused by invalid credentials")
			assert.Nil(t, result, "must not return a login object when the credentials are invalid")
		}

		{ // Make sure that we can authenticate with the new credentials.
			result, err := repo.Login(context.Background(), login.Email, newHashedPassword)
			assert.NoError(t, err, "must not return an error for valid credentials")
			assert.NotNil(t, result, "must return a login object for valid credentials")
			assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
		}
	})

	t.Run("cannot change with bad old password", func(t *testing.T) {
		login, password := fixtures.GivenIHaveLogin(t)
		hashedPassword := hash.HashPassword(login.Email, password)
		bogusHashedPassword := hash.HashPassword(login.Email, gofakeit.Generate("?????????????"))
		newHashedPassword := hash.HashPassword(login.Email, gofakeit.Generate("?????????????"))

		assert.NotEqual(t, hashedPassword, newHashedPassword, "hashed passwords must be different")

		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t))

		{ // Make sure that we can authenticate with the initial hashed password.
			result, err := repo.Login(context.Background(), login.Email, hashedPassword)
			assert.NoError(t, err, "must not return an error for valid credentials")
			assert.NotNil(t, result, "must return a login object for valid credentials")
			assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
		}

		{ // Try to update the login's password with a bogus old password. This will fail.
			err := repo.ChangePassword(context.Background(), login.LoginId, bogusHashedPassword, newHashedPassword)
			assert.EqualError(t, err, "invalid credentials provided")
			assert.Equal(t, repository.ErrInvalidCredentials, errors.Cause(err), "must be caused by invalid credentials")
		}

		{ // Make sure that we cannot authenticate using the new password we tried to change it to.
			result, err := repo.Login(context.Background(), login.Email, newHashedPassword)
			assert.EqualError(t, err, "invalid credentials provided")
			assert.Equal(t, repository.ErrInvalidCredentials, errors.Cause(err), "must be caused by invalid credentials")
			assert.Nil(t, result, "must not return a login object when the credentials are invalid")
		}

		{ // Make sure that we can still authenticate using the real old password.
			result, err := repo.Login(context.Background(), login.Email, hashedPassword)
			assert.NoError(t, err, "must not return an error for valid credentials")
			assert.NotNil(t, result, "must return a login object for valid credentials")
			assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
		}
	})

	t.Run("bad database connection", func(t *testing.T) {
		repo := repository.NewSecurityRepository(testutils.GetBadPgDatabase(t))

		email := testutils.GetUniqueEmail(t)
		bogusHashedPassword := hash.HashPassword(email, gofakeit.Generate("?????????????"))
		newHashedPassword := hash.HashPassword(email, gofakeit.Generate("?????????????"))

		err := repo.ChangePassword(context.Background(), 1234, bogusHashedPassword, newHashedPassword)
		assert.EqualError(t, err, "failed to update password: forcing a bad connection")
	})
}
