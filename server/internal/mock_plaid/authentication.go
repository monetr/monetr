package mock_plaid

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
)

const (
	RequireAccessToken      = true
	DoNotRequireAccessToken = false
)

func ValidatePlaidAuthentication(t *testing.T, request *http.Request, requireAccessToken bool) (accessToken string) {
	split := bytes.NewBuffer(nil)
	bodyReader := io.TeeReader(request.Body, split)
	request.Body = io.NopCloser(split)

	var authenticationInBody struct {
		ClientId    string `json:"client_id"`
		Secret      string `json:"secret"`
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.NewDecoder(bodyReader).Decode(&authenticationInBody), "must decode request")

	if strings.TrimSpace(authenticationInBody.ClientId) == "" {
		require.NotEmpty(t, request.Header.Get("Plaid-Client-Id"), "client Id cannot be missing")
	}

	if strings.TrimSpace(authenticationInBody.Secret) == "" {
		require.NotEmpty(t, request.Header.Get("Plaid-Secret"), "secret cannot be missing")
	}

	if requireAccessToken {
		require.NotEmpty(t, authenticationInBody.AccessToken, "access token is required")
	}

	return authenticationInBody.AccessToken
}
