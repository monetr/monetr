package repository_test

import (
	"context"
	"math"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
)

func TestSecretsRepository_Store(t *testing.T) {
	t.Run("account does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := secrets.NewPlaintextKMS()
		// user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		repo := repository.NewSecretsRepository(log, clock, db, kms, math.MaxInt64)

		err := repo.Store(context.Background(), &repository.Secret{
			Kind:   models.PlaidSecretKind,
			Secret: gofakeit.UUID(),
		})
		assert.EqualError(t, err, `failed to update access token: ERROR #23503 insert or update on table "plaid_tokens" violates foreign key constraint "fk_plaid_tokens_account"`)
	})
}
