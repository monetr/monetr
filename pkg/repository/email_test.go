package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/rest-api/pkg/internal/testutils"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestEmailRepositoryBase_SetEmailVerified(t *testing.T) {
	seedLogin := func(t *testing.T, login models.Login) {
		loginWithPassword := models.LoginWithHash{
			Login:        login,
			PasswordHash: gofakeit.Generate("?????????????????????"),
		}

		db := testutils.GetPgDatabase(t)
		result, err := db.Model(&loginWithPassword).Insert(&loginWithPassword)
		assert.NoError(t, err, "must insert login for test")
		assert.Equal(t, 1, result.RowsAffected(), "must affect 1 row for the insert")
	}

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
		db := testutils.GetPgDatabase(t)

		emailAddress := fmt.Sprintf("%s@monetr.mini", gofakeit.UUID())

		seedLogin(t, models.Login{
			Email:           emailAddress,
			FirstName:       gofakeit.FirstName(),
			LastName:        gofakeit.LastName(),
			IsEnabled:       true,
			IsEmailVerified: false,
		})

		log := testutils.GetLog(t)
		emailVerification := NewEmailRepository(log, db)

		assertEmailVerified(t, emailAddress, EmailNotVerified)

		err := emailVerification.SetEmailVerified(context.Background(), emailAddress)
		assert.NoError(t, err, "email must successfully be verified")

		assertEmailVerified(t, emailAddress, EmailVerified)
	})

	t.Run("login with email does not exist", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)

		emailAddress := fmt.Sprintf("%s@monetr.mini", gofakeit.UUID())

		log := testutils.GetLog(t)
		emailVerification := NewEmailRepository(log, db)

		err := emailVerification.SetEmailVerified(context.Background(), emailAddress)
		assert.EqualError(t, err, "email cannot be verified")
	})

	t.Run("email already verified", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)

		emailAddress := fmt.Sprintf("%s@monetr.mini", gofakeit.UUID())

		seedLogin(t, models.Login{
			Email:           emailAddress,
			FirstName:       gofakeit.FirstName(),
			LastName:        gofakeit.LastName(),
			IsEnabled:       true,
			IsEmailVerified: true,
		})

		log := testutils.GetLog(t)
		emailVerification := NewEmailRepository(log, db)

		assertEmailVerified(t, emailAddress, EmailVerified)

		err := emailVerification.SetEmailVerified(context.Background(), emailAddress)
		assert.EqualError(t, err, "email cannot be verified")

		assertEmailVerified(t, emailAddress, EmailVerified)
	})

	t.Run("login not enabled", func(t *testing.T) {
		db := testutils.GetPgDatabase(t)

		emailAddress := fmt.Sprintf("%s@monetr.mini", gofakeit.UUID())

		seedLogin(t, models.Login{
			Email:           emailAddress,
			FirstName:       gofakeit.FirstName(),
			LastName:        gofakeit.LastName(),
			IsEnabled:       false,
			IsEmailVerified: false,
		})

		log := testutils.GetLog(t)
		emailVerification := NewEmailRepository(log, db)

		assertEmailVerified(t, emailAddress, EmailNotVerified)

		err := emailVerification.SetEmailVerified(context.Background(), emailAddress)
		assert.EqualError(t, err, "email cannot be verified")

		assertEmailVerified(t, emailAddress, EmailNotVerified)
	})
}
