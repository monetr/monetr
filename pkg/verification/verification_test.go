package verification

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestVerificationBase_CreateEmailVerificationToken(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		emailRepo := repository.NewEmailRepository(log, db)
		tokens := NewJWTEmailVerification(gofakeit.Generate("????????"))

		verification := NewEmailVerification(log, time.Second, emailRepo, tokens)

		emailAddress := testutils.GetUniqueEmail(t)

		token, err := verification.CreateEmailVerificationToken(context.Background(), emailAddress)
		assert.NoError(t, err, "must be able to create email verification token without error")
		assert.NotEmpty(t, token, "token must not be empty")
	})
}
