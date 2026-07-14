package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

func TestGetFiles(t *testing.T) {
	t.Run("with a valid api key", func(t *testing.T) {
		// The files endpoint lives on the billedKeyOrToken route group so it accepts
		// an API key. It only requires authentication (plus an active subscription,
		// which passes because billing is disabled in the test config) and lists the
		// authenticated account's files. A fresh account has no files, so the handler
		// returns an empty array.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.GET(`/api/files`).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()
		response.Status(http.StatusOK)
		// The repository always returns a non-nil slice, so the body is a JSON array
		// (empty for a brand new account).
		response.JSON().Array().Empty()
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		// A syntactically plausible but non-existent API key must be rejected before
		// the handler runs.
		_, e := NewTestApplication(t)

		response := e.GET(`/api/files`).
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			Expect()
		response.Status(http.StatusUnauthorized)
	})
}
