package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

func seedLogin(ctx context.Context, t *testing.T, db bun.IDB, login *models.Login) {
	loginWithPassword := models.LoginWithHash{
		Login:        *login,
		PasswordHash: gofakeit.Generate("?????????????????????"),
	}

	result, err := db.NewInsert().Model(&loginWithPassword).Exec(ctx, &loginWithPassword)
	assert.NoError(t, err, "must insert login for test")
	affected, err := result.RowsAffected()
	assert.NoError(t, err, "must retrieve rows affected")
	assert.Equal(t, 1, affected, "must affect 1 row for the insert")

	login.LoginId = loginWithPassword.LoginId
}

func TestEmailRepositoryBase_SetEmailVerified(t *testing.T) {
	assertEmailVerified := func(ctx context.Context, t *testing.T, db bun.IDB, emailAddress string, verified bool) {
		exists, err := db.NewSelect().
			Model(&models.Login{}).
			Where(`login.email = ?`, strings.ToLower(emailAddress)).
			Where(`login.is_email_verified = ?`, verified).
			Exists(ctx)
		assert.NoError(t, err, "must assert that the email is verified")
		assert.True(t, exists, "login must be in the expected state")
	}

	t.Run("happy path", func(t *testing.T) {
		testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			emailAddress := testutils.GetUniqueEmail(t)

			seedLogin(ctx, t, db, &models.Login{
				Email:           emailAddress,
				FirstName:       gofakeit.FirstName(),
				LastName:        gofakeit.LastName(),
				IsEnabled:       true,
				IsEmailVerified: false,
			})

			log := testutils.GetLog(t)
			emailVerification := NewEmailRepository(log, db)

			assertEmailVerified(ctx, t, db, emailAddress, EmailNotVerified)

			err := emailVerification.SetEmailVerified(ctx, emailAddress)
			assert.NoError(t, err, "email must successfully be verified")

			assertEmailVerified(ctx, t, db, emailAddress, EmailVerified)
		})
	})

	t.Run("login with email does not exist", func(t *testing.T) {
		testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			emailAddress := testutils.GetUniqueEmail(t)

			log := testutils.GetLog(t)
			emailVerification := NewEmailRepository(log, db)

			err := emailVerification.SetEmailVerified(ctx, emailAddress)
			assert.EqualError(t, err, "email cannot be verified")
		})
	})

	t.Run("email already verified", func(t *testing.T) {
		testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			emailAddress := testutils.GetUniqueEmail(t)

			seedLogin(ctx, t, db, &models.Login{
				Email:           emailAddress,
				FirstName:       gofakeit.FirstName(),
				LastName:        gofakeit.LastName(),
				IsEnabled:       true,
				IsEmailVerified: true,
			})

			log := testutils.GetLog(t)
			emailVerification := NewEmailRepository(log, db)

			assertEmailVerified(ctx, t, db, emailAddress, EmailVerified)

			err := emailVerification.SetEmailVerified(ctx, emailAddress)
			assert.EqualError(t, err, "email cannot be verified")

			assertEmailVerified(ctx, t, db, emailAddress, EmailVerified)
		})
	})

	t.Run("login not enabled", func(t *testing.T) {
		testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			emailAddress := testutils.GetUniqueEmail(t)

			seedLogin(ctx, t, db, &models.Login{
				Email:           emailAddress,
				FirstName:       gofakeit.FirstName(),
				LastName:        gofakeit.LastName(),
				IsEnabled:       false,
				IsEmailVerified: false,
			})

			log := testutils.GetLog(t)
			emailVerification := NewEmailRepository(log, db)

			assertEmailVerified(ctx, t, db, emailAddress, EmailNotVerified)

			err := emailVerification.SetEmailVerified(ctx, emailAddress)
			assert.EqualError(t, err, "email cannot be verified")

			assertEmailVerified(ctx, t, db, emailAddress, EmailNotVerified)
		})
	})
}

func TestEmailRepositoryBase_GetLoginForEmail(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			emailAddress := testutils.GetUniqueEmail(t)

			originalLogin := &models.Login{
				Email:     emailAddress,
				FirstName: gofakeit.FirstName(),
				LastName:  gofakeit.LastName(),
			}
			seedLogin(ctx, t, db, originalLogin)

			log := testutils.GetLog(t)
			emailVerification := NewEmailRepository(log, db)

			login, err := emailVerification.GetLoginForEmail(ctx, emailAddress)
			assert.NoError(t, err, "must retrieve login for email successfully")
			assert.NotNil(t, login, "login result should not be nil")
			assert.Equal(t, originalLogin.LoginId, login.LoginId, "login Id should match expected")
		})
	})

	t.Run("email does not exist", func(t *testing.T) {
		testutils.ForEachDatabase(t, func(ctx context.Context, t *testing.T, db *bun.DB) {
			emailAddress := testutils.GetUniqueEmail(t)

			log := testutils.GetLog(t)
			emailVerification := NewEmailRepository(log, db)

			login, err := emailVerification.GetLoginForEmail(ctx, emailAddress)
			assert.EqualError(t, err, "failed to retrieve login by email: pg: no rows in result set")
			assert.Nil(t, login, "login result should be nil")
		})
	})
}
