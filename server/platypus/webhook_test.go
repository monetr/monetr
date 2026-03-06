package platypus

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/mock_plaid"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/secrets"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryWebhookVerification(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		clock := clock.NewMock()
		mock_plaid.MockGetWebhookVerificationKey(t, clock, kid, &privateKey.PublicKey)
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabaseTxn(t)
		kms := secrets.NewPlaintextKMS()

		plaid := NewPlaid(log, clock, kms, db, config.Plaid{
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		})

		webhookVerification := NewInMemoryWebhookVerification(log, plaid, time.Second*1)

		verify, err := webhookVerification.GetVerificationKey(context.Background(), kid)
		assert.NoError(t, err, "must get verification")
		assert.NotNil(t, verify, "verify must not be nil")
	})
}
