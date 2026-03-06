package controller_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/controller"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_plaid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPlaidWebhook(t *testing.T) {
	t.Run("not found when not enabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = false
		_, e := NewTestApplicationWithConfig(t, config)

		response := e.POST(`/api/plaid/webhook`).
			WithJSON(controller.PlaidWebhook{
				WebhookType:     "TRANSACTIONS",
				WebhookCode:     "DEFAULT_UPDATE",
				ItemId:          gofakeit.UUID(),
				NewTransactions: 3,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("plaid webhooks are not enabled")
	})

	t.Run("not authorized", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		_, e := NewTestApplicationWithConfig(t, config)

		response := e.POST(`/api/plaid/webhook`).
			WithJSON(controller.PlaidWebhook{
				WebhookType:     "TRANSACTIONS",
				WebhookCode:     "DEFAULT_UPDATE",
				ItemId:          gofakeit.UUID(),
				NewTransactions: 3,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("bad authorization", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		_, e := NewTestApplicationWithConfig(t, config)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", "abc123").
			WithJSON(controller.PlaidWebhook{
				WebhookType:     "TRANSACTIONS",
				WebhookCode:     "DEFAULT_UPDATE",
				ItemId:          gofakeit.UUID(),
				NewTransactions: 3,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
	})

	t.Run("successful transaction sync webhook", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		app, e := NewTestApplicationWithConfig(t, config)

		user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKey(
			t,
			app.Clock,
			kid,
			&privateKey.PublicKey,
		)

		jwt.TimeFunc = app.Clock.Now
		t.Cleanup(func() { jwt.TimeFunc = time.Now })

		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
		)

		app.Jobs.EXPECT().
			EnqueueJob(gomock.Any(), background.SyncPlaid, gomock.Any()).
			Return(nil).
			Times(1)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(controller.PlaidWebhook{
				WebhookType: "TRANSACTIONS",
				WebhookCode: "SYNC_UPDATES_AVAILABLE",
				ItemId:      link.PlaidLink.PlaidId,
			}).
			Expect()

		response.Status(http.StatusOK)
		assert.EqualValues(t, map[string]int{
			"POST https://sandbox.plaid.com/webhook_verification_key/get": 1,
		}, httpmock.GetCallCountInfo(), "must match expected Plaid API calls")
	})

	t.Run("won't re-request the same key ID", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		app, e := NewTestApplicationWithConfig(t, config)

		// Because the in memory webhook code does not use a mock clock.
		app.Clock.Add(time.Now().Sub(app.Clock.Now()))

		user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKey(
			t,
			app.Clock,
			kid,
			&privateKey.PublicKey,
		)

		jwt.TimeFunc = app.Clock.Now
		t.Cleanup(func() { jwt.TimeFunc = time.Now })

		app.Jobs.EXPECT().
			EnqueueJob(gomock.Any(), background.SyncPlaid, background.SyncPlaidArguments{
				AccountId: link.AccountId,
				LinkId:    link.LinkId,
				Trigger:   "webhook",
			}).
			Return(nil).
			Times(2)

		{ // First request goes through and should succeed
			// Generate a verification token right now, using the key ID that we have
			// picked.
			verificationToken := mock_plaid.SignWebhookJWT(
				t,
				app.Clock,
				privateKey,
				kid,
			)
			response := e.POST(`/api/plaid/webhook`).
				WithHeader("Plaid-Verification", verificationToken).
				WithJSON(controller.PlaidWebhook{
					WebhookType: "TRANSACTIONS",
					WebhookCode: "SYNC_UPDATES_AVAILABLE",
					ItemId:      link.PlaidLink.PlaidId,
				}).
				Expect()

			response.Status(http.StatusOK)
		}

		// Progress time forward a few seconds so the next verification key is a bit
		// different.
		app.Clock.Add(5 * time.Second)

		{ // First request goes through and should succeed
			// Generate another verification key using the new timestamp but the same
			// Key ID. We shouldn't request the key again from Plaid which we will
			// assert at the bottom!
			verificationToken := mock_plaid.SignWebhookJWT(
				t,
				app.Clock,
				privateKey,
				kid,
			)
			response := e.POST(`/api/plaid/webhook`).
				WithHeader("Plaid-Verification", verificationToken).
				WithJSON(controller.PlaidWebhook{
					WebhookType: "TRANSACTIONS",
					WebhookCode: "SYNC_UPDATES_AVAILABLE",
					ItemId:      link.PlaidLink.PlaidId,
				}).
				Expect()

			response.Status(http.StatusOK)
		}

		assert.EqualValues(t, map[string]int{
			"POST https://sandbox.plaid.com/webhook_verification_key/get": 1,
		}, httpmock.GetCallCountInfo(), "must match expected Plaid API calls")
	})

	t.Run("successful item error webhook updates link status", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		app, e := NewTestApplicationWithConfig(t, config)

		user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKey(
			t,
			app.Clock,
			kid,
			&privateKey.PublicKey,
		)

		jwt.TimeFunc = app.Clock.Now
		t.Cleanup(func() { jwt.TimeFunc = time.Now })

		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
		)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(controller.PlaidWebhook{
				WebhookType: "ITEM",
				WebhookCode: "ERROR",
				ItemId:      link.PlaidLink.PlaidId,
				Error:       map[string]any{"error_code": "ITEM_LOGIN_REQUIRED"},
			}).
			Expect()

		response.Status(http.StatusOK)
		assert.EqualValues(t, map[string]int{
			"POST https://sandbox.plaid.com/webhook_verification_key/get": 1,
		}, httpmock.GetCallCountInfo(), "must match expected Plaid API calls")
	})

	t.Run("expired JWT is rejected", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		app, e := NewTestApplicationWithConfig(t, config)

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKey(
			t,
			app.Clock,
			kid,
			&privateKey.PublicKey,
		)

		jwt.TimeFunc = app.Clock.Now
		t.Cleanup(func() { jwt.TimeFunc = time.Now })

		claims := controller.PlaidClaims{
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  app.Clock.Now().Add(-10 * time.Minute).Unix(),
				ExpiresAt: app.Clock.Now().Add(-5 * time.Minute).Unix(),
			},
		}
		expiredToken := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		expiredToken.Header["kid"] = kid
		verificationToken, err := expiredToken.SignedString(privateKey)
		require.NoError(t, err, "must sign expired webhook JWT")

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(controller.PlaidWebhook{
				WebhookType: "TRANSACTIONS",
				WebhookCode: "SYNC_UPDATES_AVAILABLE",
				ItemId:      gofakeit.UUID(),
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
		assert.EqualValues(t, map[string]int{
			"POST https://sandbox.plaid.com/webhook_verification_key/get": 1,
		}, httpmock.GetCallCountInfo(), "must match expected Plaid API calls")
	})

	t.Run("failed JWK retrieval is rejected", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		app, e := NewTestApplicationWithConfig(t, config)

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKeyFailure(t)

		jwt.TimeFunc = app.Clock.Now
		t.Cleanup(func() { jwt.TimeFunc = time.Now })

		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
		)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(controller.PlaidWebhook{
				WebhookType: "TRANSACTIONS",
				WebhookCode: "SYNC_UPDATES_AVAILABLE",
				ItemId:      gofakeit.UUID(),
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
		assert.EqualValues(t, map[string]int{
			"POST https://sandbox.plaid.com/webhook_verification_key/get": 1,
		}, httpmock.GetCallCountInfo(), "must match expected Plaid API calls")
	})

	t.Run("valid JWT for unknown item returns 200", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		app, e := NewTestApplicationWithConfig(t, config)

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKey(
			t,
			app.Clock,
			kid,
			&privateKey.PublicKey,
		)

		jwt.TimeFunc = app.Clock.Now
		t.Cleanup(func() { jwt.TimeFunc = time.Now })

		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
		)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(controller.PlaidWebhook{
				WebhookType: "TRANSACTIONS",
				WebhookCode: "SYNC_UPDATES_AVAILABLE",
				ItemId:      gofakeit.UUID(),
			}).
			Expect()

		response.Status(http.StatusOK)
		assert.EqualValues(t, map[string]int{
			"POST https://sandbox.plaid.com/webhook_verification_key/get": 1,
		}, httpmock.GetCallCountInfo(), "must match expected Plaid API calls")
	})
}
