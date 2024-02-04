package mock_teller

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	RequireAccessToken      = true
	DoNotRequireAccessToken = false
)

func ValidateTellerAuthentication(t *testing.T, request *http.Request, requireAccessToken bool) (accessToken string) {
	authHeader := request.Header.Get("Authorization")
	if requireAccessToken {
		require.NotEmpty(t, authHeader, "Teller authentication must be present!")
		require.True(t, strings.HasPrefix(authHeader, "Basic "), "must have Basic auth prefix")
		auth := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(auth)
		require.NoError(t, err, "authentication must be proper base64!")
		credentials := strings.Split(string(decoded), ":")
		require.Len(t, credentials, 2, "must have 2 parts to the authentication header!")

		return credentials[0]
	}

	require.Empty(t, authHeader, "auth header should not be present for this request!")

	return ""
}
