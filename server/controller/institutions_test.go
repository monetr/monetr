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
}
