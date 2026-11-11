package controller_test

import (
	"net/http"
	"testing"
)

func TestGetInstitutionDetails(t *testing.T) {
	t.Run("does not accept an api key", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		// The institution details endpoint is backed by Plaid and is intentionally
		// token only, a valid API key must not be accepted as authentication for
		// it.
		e.GET(`/api/institutions/{institutionId}`).
			WithPath("institutionId", "ins_109508").
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect().
			Status(http.StatusUnauthorized)
	})

	t.Run("plaid is not enabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Plaid.Enabled = false
		_, e := NewTestApplicationWithConfig(t, config)
		token := GivenIHaveToken(t, e)

		response := e.GET(`/api/institutions/{institutionId}`).
			WithPath("institutionId", "ins_109508").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("Plaid is not enabled on this server")
	})
}
