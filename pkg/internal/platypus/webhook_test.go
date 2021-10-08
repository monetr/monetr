package platypus

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/mock_plaid"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewInMemoryWebhookVerification(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock_plaid.MockGetWebhookVerificationKey(t)

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabaseTxn(t)
		secret := secrets.NewPostgresPlaidSecretsProvider(log, db)
		plaidRepo := repository.NewPlaidRepository(db)

		plaid := NewPlaid(log, secret, plaidRepo, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		webhookVerification := NewInMemoryWebhookVerification(log, plaid, time.Second * 1)

		verify, err := webhookVerification.GetVerificationKey(context.Background(), gofakeit.UUID())
		assert.NoError(t, err, "must get verification")
		assert.NotNil(t, verify, "verify must not be nil")
	})
}
