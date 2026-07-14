package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

func TestDeleteAccount(t *testing.T) {
	t.Run("with a valid api key", func(t *testing.T) {
		// The delete account endpoint lives on the billedKeyOrToken route group so it
		// accepts an API key. The deleteAccount handler is currently a stub that
		// always returns 501 Not Implemented. A valid API key still AUTHENTICATES and
		// reaches the handler, so the expected status for a valid key is 501 (not
		// 401). This proves the key was accepted and the request made it through the
		// auth and subscription middleware into the handler.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.DELETE(`/api/account`).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()
		response.Status(http.StatusNotImplemented)
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		// A syntactically plausible but non-existent API key must be rejected before
		// the handler runs.
		_, e := NewTestApplication(t)

		response := e.DELETE(`/api/account`).
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			Expect()
		response.Status(http.StatusUnauthorized)
	})
}
