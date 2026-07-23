//go:build icons && simple_icons

package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

func TestSearchIcon(t *testing.T) {
	t.Run("get amazon icon", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.POST("/api/icons/search").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "amazon",
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.svg").String().NotEmpty()
	})

	t.Run("with a valid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		// /icons/search accepts an API key as authentication, provided via the
		// basic auth header where the username is the key Id and the password is
		// the secret.
		response := e.POST("/api/icons/search").
			WithBasicAuth(apiKeyId, apiKeySecret).
			WithJSON(map[string]any{
				"name": "amazon",
			}).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.svg").String().NotEmpty()
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		// A well formed but unknown API key must be rejected.
		response := e.POST("/api/icons/search").
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			WithJSON(map[string]any{
				"name": "amazon",
			}).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}
