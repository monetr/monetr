package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/controller"
)

func TestPlaidWebhook(t *testing.T) {
	t.Run("not found when not enabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = false
		e := NewTestApplicationWithConfig(t, config)

		response := e.POST(`/api/plaid/webhook`).
			WithJSON(controller.PlaidWebhook{
				WebhookType:     "TRANSACTIONS",
				WebhookCode:     "DEFAULT_UPDATE",
				ItemId:          gofakeit.UUID(),
				NewTransactions: 3,
			}).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().Equal("the requested path does not exist")
	})

	t.Run("not authorized", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		e := NewTestApplicationWithConfig(t, config)

		response := e.POST(`/api/plaid/webhook`).
			WithJSON(controller.PlaidWebhook{
				WebhookType:     "TRANSACTIONS",
				WebhookCode:     "DEFAULT_UPDATE",
				ItemId:          gofakeit.UUID(),
				NewTransactions: 3,
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
		response.JSON().Path("$.error").String().Equal("unauthorized")
	})

	t.Run("bad authorization", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Plaid.WebhooksEnabled = true
		e := NewTestApplicationWithConfig(t, config)

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
		response.JSON().Path("$.error").String().Equal("unauthorized: token contains an invalid number of segments")
	})
}
