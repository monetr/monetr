package controller_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/controller"
	"github.com/monetr/monetr/server/datasources/plaid/plaid_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_plaid"
	"github.com/monetr/monetr/server/internal/mockqueue"
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

		// Because the in memory webhook code does not use a mock clock.
		app.Clock.Add(time.Now().Sub(app.Clock.Now()))

		user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)

		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(plaid_jobs.SyncPlaid),
				gomock.Any(),
				gomock.Eq(plaid_jobs.SyncPlaidArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
					Trigger:   "webhook",
				}),
			).
			Times(1).
			Return(nil)

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKey(
			t,
			app.Clock,
			kid,
			&privateKey.PublicKey,
		)

		body := controller.PlaidWebhook{
			WebhookType: "TRANSACTIONS",
			WebhookCode: "SYNC_UPDATES_AVAILABLE",
			ItemId:      link.PlaidLink.PlaidId,
		}
		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
			body,
		)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(body).
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

		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue).
			Times(2)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(plaid_jobs.SyncPlaid),
				gomock.Any(),
				gomock.Eq(plaid_jobs.SyncPlaidArguments{
					AccountId: link.AccountId,
					LinkId:    link.LinkId,
					Trigger:   "webhook",
				}),
			).
			Times(2).
			Return(nil)

		{ // First request goes through and should succeed
			// Generate a verification token right now, using the key ID that we have
			// picked.
			body := controller.PlaidWebhook{
				WebhookType: "TRANSACTIONS",
				WebhookCode: "SYNC_UPDATES_AVAILABLE",
				ItemId:      link.PlaidLink.PlaidId,
			}
			verificationToken := mock_plaid.SignWebhookJWT(
				t,
				app.Clock,
				privateKey,
				kid,
				body,
			)
			response := e.POST(`/api/plaid/webhook`).
				WithHeader("Plaid-Verification", verificationToken).
				WithJSON(body).
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
			body := controller.PlaidWebhook{
				WebhookType: "TRANSACTIONS",
				WebhookCode: "SYNC_UPDATES_AVAILABLE",
				ItemId:      link.PlaidLink.PlaidId,
			}
			verificationToken := mock_plaid.SignWebhookJWT(
				t,
				app.Clock,
				privateKey,
				kid,
				body,
			)
			response := e.POST(`/api/plaid/webhook`).
				WithHeader("Plaid-Verification", verificationToken).
				WithJSON(body).
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

		body := controller.PlaidWebhook{
			WebhookType: "ITEM",
			WebhookCode: "ERROR",
			ItemId:      link.PlaidLink.PlaidId,
			Error:       map[string]any{"error_code": "ITEM_LOGIN_REQUIRED"},
		}
		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
			body,
		)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(body).
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

		// Because the in memory webhook code does not use a mock clock.
		app.Clock.Add(time.Now().Sub(app.Clock.Now()))

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKey(
			t,
			app.Clock,
			kid,
			&privateKey.PublicKey,
		)

		claims := controller.PlaidClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				IssuedAt:  jwt.NewNumericDate(app.Clock.Now().Add(-10 * time.Minute)),
				ExpiresAt: jwt.NewNumericDate(app.Clock.Now().Add(-5 * time.Minute)),
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

		// Because the in memory webhook code does not use a mock clock.
		app.Clock.Add(time.Now().Sub(app.Clock.Now()))

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKeyFailure(t)

		body := controller.PlaidWebhook{
			WebhookType: "TRANSACTIONS",
			WebhookCode: "SYNC_UPDATES_AVAILABLE",
			ItemId:      gofakeit.UUID(),
		}
		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
			body,
		)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(body).
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

		// Because the in memory webhook code does not use a mock clock.
		app.Clock.Add(time.Now().Sub(app.Clock.Now()))

		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		require.NoError(t, err, "must generate EC key")

		kid := gofakeit.UUID()
		mock_plaid.MockGetWebhookVerificationKey(
			t,
			app.Clock,
			kid,
			&privateKey.PublicKey,
		)

		body := controller.PlaidWebhook{
			WebhookType: "TRANSACTIONS",
			WebhookCode: "SYNC_UPDATES_AVAILABLE",
			ItemId:      gofakeit.UUID(),
		}
		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
			body,
		)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(body).
			Expect()

		response.Status(http.StatusOK)
		assert.EqualValues(t, map[string]int{
			"POST https://sandbox.plaid.com/webhook_verification_key/get": 1,
		}, httpmock.GetCallCountInfo(), "must match expected Plaid API calls")
	})

	t.Run("fails if the signature doesn't match", func(t *testing.T) {
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

		verificationToken := mock_plaid.SignWebhookJWT(
			t,
			app.Clock,
			privateKey,
			kid,
			controller.PlaidWebhook{
				WebhookType: "TRANSACTIONS",
				// Force the signature to not match so that way we reject it, even if it
				// is a valid JWT.
				WebhookCode: "LITERALLY ANYTHING",
				ItemId:      link.PlaidLink.PlaidId,
			},
		)

		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(plaid_jobs.SyncPlaid),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0).
			Return(nil)

		response := e.POST(`/api/plaid/webhook`).
			WithHeader("Plaid-Verification", verificationToken).
			WithJSON(controller.PlaidWebhook{
				WebhookType: "TRANSACTIONS",
				WebhookCode: "SYNC_UPDATES_AVAILABLE",
				ItemId:      link.PlaidLink.PlaidId,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().IsEqual("unauthorized")
		assert.EqualValues(t, map[string]int{
			"POST https://sandbox.plaid.com/webhook_verification_key/get": 1,
		}, httpmock.GetCallCountInfo(), "must match expected Plaid API calls")
	})
}
