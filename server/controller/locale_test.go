package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

func TestListCurrencies(t *testing.T) {
	t.Run("with a valid api key", func(t *testing.T) {
		// The locale/currency endpoint lives on the billedKeyOrToken route group so
		// it accepts an API key. It only requires authentication (plus an active
		// subscription, which passes because billing is disabled in the test
		// config), so a valid key should reach the handler and return the installed
		// currency list.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.GET(`/api/locale/currency`).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()
		response.Status(http.StatusOK)
		// The handler returns the list of installed ISO currency codes as a JSON
		// array, assert the shape is present.
		response.JSON().Array().NotEmpty()
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		// A syntactically plausible but non-existent API key must be rejected before
		// the handler runs.
		_, e := NewTestApplication(t)

		response := e.GET(`/api/locale/currency`).
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			Expect()
		response.Status(http.StatusUnauthorized)
	})
}
