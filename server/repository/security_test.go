package repository_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBaseSecurityRepository_Login(t *testing.T) {
	t.Run("valid credentials", func(t *testing.T) {
		clock := clock.NewMock()
		login, password := fixtures.GivenIHaveLogin(t, clock)
		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t), clock)

		result, _, err := repo.Login(context.Background(), login.Email, password)
		assert.NoError(t, err, "must not return an error for valid credentials")
		assert.NotNil(t, result, "must return a login object for valid credentials")
		assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
	})

	t.Run("oddly cased email", func(t *testing.T) {
		clock := clock.NewMock()
		login, password := fixtures.GivenIHaveLogin(t, clock)
		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t), clock)

		email := strings.ToUpper(login.Email)

		result, _, err := repo.Login(context.Background(), email, password)
		assert.NoError(t, err, "must not return an error for valid credentials")
		assert.NotNil(t, result, "must return a login object for valid credentials")
		assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
	})

	t.Run("invalid credentials", func(t *testing.T) {
		clock := clock.NewMock()
		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t), clock)
		email := testutils.GetUniqueEmail(t)
		password := gofakeit.Generate("????????")

		result, _, err := repo.Login(context.Background(), email, password)
		assert.EqualError(t, err, "invalid credentials provided")
		assert.Equal(t, repository.ErrInvalidCredentials, errors.Cause(err), "must be caused by invalid credentials")
		assert.Nil(t, result, "must not return a login object when the credentials are invalid")
	})

	t.Run("bad database connection", func(t *testing.T) {
		clock := clock.NewMock()
		login, password := fixtures.GivenIHaveLogin(t, clock)
		repo := repository.NewSecurityRepository(testutils.GetBadPgDatabase(t), clock)

		result, _, err := repo.Login(context.Background(), login.Email, password)
		assert.EqualError(t, err, "failed to verify credentials: forcing a bad connection")
		assert.Nil(t, result, "must not return a result if the connection is bad")
	})
}

func TestBaseSecurityRepository_ChangePassword(t *testing.T) {
	t.Run("successful", func(t *testing.T) {
		clock := clock.NewMock()
		login, password := fixtures.GivenIHaveLogin(t, clock)
		newPassword := gofakeit.Generate("?????????????")

		assert.NotEqual(t, password, newPassword, "passwords must be different")

		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t), clock)

		{ // Make sure that we can authenticate with the initial hashed password.
			result, _, err := repo.Login(context.Background(), login.Email, password)
			assert.NoError(t, err, "must not return an error for valid credentials")
			assert.NotNil(t, result, "must return a login object for valid credentials")
			assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
		}

		{ // Update the login's password.
			err := repo.ChangePassword(context.Background(), login.LoginId, password, newPassword)
			assert.NoError(t, err, "must not return an error when changing the password")
		}

		{ // Make sure that we can no longer authenticate using the old credentials.
			result, _, err := repo.Login(context.Background(), login.Email, password)
			assert.EqualError(t, err, "invalid credentials provided")
			assert.Equal(t, repository.ErrInvalidCredentials, errors.Cause(err), "must be caused by invalid credentials")
			assert.Nil(t, result, "must not return a login object when the credentials are invalid")
		}

		{ // Make sure that we can authenticate with the new credentials.
			result, _, err := repo.Login(context.Background(), login.Email, newPassword)
			assert.NoError(t, err, "must not return an error for valid credentials")
			assert.NotNil(t, result, "must return a login object for valid credentials")
			assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
		}
	})

	t.Run("cannot change with bad old password", func(t *testing.T) {
		clock := clock.NewMock()
		login, password := fixtures.GivenIHaveLogin(t, clock)
		bogusPassword := gofakeit.Generate("?????????????")
		assert.NotEqual(t, password, bogusPassword, "bogus password cannot match the real one")
		newPassword := gofakeit.Generate("?????????????")
		assert.NotEqual(t, password, newPassword, "new password cannot match the current password")

		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t), clock)

		{ // Make sure that we can authenticate with the initial hashed password.
			result, _, err := repo.Login(context.Background(), login.Email, password)
			assert.NoError(t, err, "must not return an error for valid credentials")
			assert.NotNil(t, result, "must return a login object for valid credentials")
			assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
		}

		{ // Try to update the login's password with a bogus old password. This will fail.
			err := repo.ChangePassword(context.Background(), login.LoginId, bogusPassword, newPassword)
			assert.EqualError(t, err, "invalid credentials provided")
			assert.Equal(t, repository.ErrInvalidCredentials, errors.Cause(err), "must be caused by invalid credentials")
		}

		{ // Make sure that we cannot authenticate using the new password we tried to change it to.
			result, _, err := repo.Login(context.Background(), login.Email, newPassword)
			assert.EqualError(t, err, "invalid credentials provided")
			assert.Equal(t, repository.ErrInvalidCredentials, errors.Cause(err), "must be caused by invalid credentials")
			assert.Nil(t, result, "must not return a login object when the credentials are invalid")
		}

		{ // Make sure that we can still authenticate using the real old password.
			result, _, err := repo.Login(context.Background(), login.Email, password)
			assert.NoError(t, err, "must not return an error for valid credentials")
			assert.NotNil(t, result, "must return a login object for valid credentials")
			assert.Equal(t, login.LoginId, result.LoginId, "must return the same login as the fixture")
		}
	})

	t.Run("bad database connection", func(t *testing.T) {
		clock := clock.NewMock()
		repo := repository.NewSecurityRepository(testutils.GetBadPgDatabase(t), clock)
		bogusPassword := gofakeit.Generate("?????????????")
		newPassword := gofakeit.Generate("?????????????")

		err := repo.ChangePassword(context.Background(), "user_bogus", bogusPassword, newPassword)
		assert.EqualError(t, err, "failed to find login record to change password: forcing a bad connection")
	})
}

func TestBaseSecurityRepository_SetupTOTP(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		assert.Empty(t, login.TOTP, "login should not have a TOTP secret initially")
		assert.Empty(t, login.TOTPRecoveryCodes, "login should have no TOTP recovery codes")
		assert.Nil(t, login.TOTPEnabledAt, "login should have a nil TOTP enabled at")

		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t), clock)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		uri, recoveryCodes, err := repo.SetupTOTP(ctx, login.LoginId)
		assert.NoError(t, err, "should setup TOTP without an error")
		assert.Len(t, recoveryCodes, 10, "should return 10 recovery codes")
		assert.NotEmpty(t, uri, "should return a TOTP uri")

		afterSetup := testutils.MustDBRead(t, login)
		assert.NotEmpty(t, afterSetup.TOTP, "login should have a TOTP secret after setup")
		assert.EqualValues(t, recoveryCodes, afterSetup.TOTPRecoveryCodes, "login should have the returned TOTP recovery codes")
		assert.Nil(t, afterSetup.TOTPEnabledAt, "TOTP enabled at should still be nil")
	})

	t.Run("already has TOTP enabled", func(t *testing.T) {
		clock := clock.NewMock()
		login, _ := fixtures.GivenIHaveLogin(t, clock)
		repo := repository.NewSecurityRepository(testutils.GetPgDatabase(t), clock)

		{ // Initially setup the TOTP stuff.
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_, _, err := repo.SetupTOTP(ctx, login.LoginId)
			assert.NoError(t, err, "should setup TOTP without an error")
		}

		{ // Enable the TOTP
			login.TOTPEnabledAt = myownsanity.TimeP(clock.Now())
			testutils.MustDBUpdate(t, &login)
		}

		// Now try to setup TOTP again, we should get an error.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		uri, recoveryCodes, err := repo.SetupTOTP(ctx, login.LoginId)
		assert.EqualError(t, err, "login already has TOTP enabled")
		assert.Empty(t, recoveryCodes, "should not return any recovery codes")
		assert.Empty(t, uri, "should not return a TOTP URI")
	})
}
