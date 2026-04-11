package controller_test

import (
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stripe/stripe-go/v81/webhook"
)

func TestHandleStripeWebhook(t *testing.T) {
	t.Run("not found when stripe is not enabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Stripe.Enabled = false
		config.Stripe.WebhooksEnabled = true
		_, e := NewTestApplicationWithConfig(t, config)

		response := e.POST(`/api/stripe/webhook`).
			WithHeader("Stripe-Signature", "t=0,v1=00").
			WithBytes([]byte(`{}`)).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("stripe webhooks not enabled on this server")
	})

	t.Run("not found when stripe webhooks are not enabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Stripe.Enabled = true
		config.Stripe.APIKey = gofakeit.UUID()
		config.Stripe.WebhooksEnabled = false
		_, e := NewTestApplicationWithConfig(t, config)

		response := e.POST(`/api/stripe/webhook`).
			WithHeader("Stripe-Signature", "t=0,v1=00").
			WithBytes([]byte(`{}`)).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("stripe webhooks not enabled on this server")
	})

	t.Run("bad request when stripe signature is missing", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Stripe.Enabled = true
		config.Stripe.APIKey = gofakeit.UUID()
		config.Stripe.WebhooksEnabled = true
		config.Stripe.WebhookSecret = "whsec_" + gofakeit.UUID()
		_, e := NewTestApplicationWithConfig(t, config)

		response := e.POST(`/api/stripe/webhook`).
			WithBytes([]byte(`{}`)).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("stripe signature is missing")
	})

	t.Run("successful webhook with a small body", func(t *testing.T) {
		webhookSecret := "whsec_" + gofakeit.UUID()

		config := NewTestApplicationConfig(t)
		config.Stripe.Enabled = true
		config.Stripe.APIKey = gofakeit.UUID()
		config.Stripe.WebhooksEnabled = true
		config.Stripe.WebhookSecret = webhookSecret
		_, e := NewTestApplicationWithConfig(t, config)

		// Use an event type that the billing webhook does not handle so that we can
		// exercise the controller without standing up additional billing fixtures.
		body := []byte(`{"id":"evt_test","type":"unknown.event"}`)
		signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
			Payload: body,
			Secret:  webhookSecret,
		})

		response := e.POST(`/api/stripe/webhook`).
			WithHeader("Stripe-Signature", signed.Header).
			WithBytes(body).
			Expect()

		response.Status(http.StatusOK)
	})

	t.Run("rejects body larger than the read limit", func(t *testing.T) {
		webhookSecret := "whsec_" + gofakeit.UUID()

		config := NewTestApplicationConfig(t)
		config.Stripe.Enabled = true
		config.Stripe.APIKey = gofakeit.UUID()
		config.Stripe.WebhooksEnabled = true
		config.Stripe.WebhookSecret = webhookSecret
		_, e := NewTestApplicationWithConfig(t, config)

		// Build a payload that exceeds the 65535 byte limit on the stripe webhook
		// handler. We sign the entire body so the signature is valid against the
		// unbounded payload, but the controller now caps the read at math.MaxUint16
		// bytes which means the bytes the server actually reads cannot match the
		// signature we generated. The controller must reject the request rather
		// than allowing an unbounded body to be read into memory. Combined with the
		// "successful webhook with a small body" subtest above, this proves the
		// limit is in effect: a payload that would otherwise have a valid signature
		// is rejected solely because the read was truncated.
		padding := strings.Repeat(" ", math.MaxUint16+1)
		body := []byte(`{"id":"evt_test","type":"unknown.event","description":"` + padding + `"}`)

		signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
			Payload: body,
			Secret:  webhookSecret,
		})

		response := e.POST(`/api/stripe/webhook`).
			WithHeader("Stripe-Signature", signed.Header).
			WithBytes(body).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("failed to validate stripe event")
	})
}
