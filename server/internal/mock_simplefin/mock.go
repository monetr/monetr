package mock_simplefin

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// TestUsername and TestPassword are the only credentials the mock accounts
	// endpoint will accept. Anything else gets a 403 like the real Bridge.
	TestUsername = "monetr-test-user"
	TestPassword = "monetr-test-password-that-should-never-be-logged"
)

func Path(t *testing.T, relative string) string {
	require.NotEmpty(t, relative, "relative url cannot be empty")
	return (&url.URL{
		Scheme: "https",
		Host:   config.DefaultSimpleFinHost,
		Path:   relative,
	}).String()
}

// AccessUrl returns an access URL with the test credentials in it, the same
// shape as what the Bridge gives back when a setup token is claimed.
func AccessUrl(t *testing.T) string {
	return (&url.URL{
		Scheme: "https",
		User:   url.UserPassword(TestUsername, TestPassword),
		Host:   config.DefaultSimpleFinHost,
		Path:   "/simplefin",
	}).String()
}

// SetupToken returns the setup token for the provided claim URL. Real setup
// tokens are just the claim URL base64 encoded.
func SetupToken(t *testing.T, claimUrl string) string {
	require.NotEmpty(t, claimUrl, "claim url cannot be empty")
	return base64.StdEncoding.EncodeToString([]byte(claimUrl))
}

func ValidateSimpleFinAuthentication(t *testing.T, request *http.Request) bool {
	username, password, ok := request.BasicAuth()
	assert.True(t, ok, "request must use basic auth")
	assert.Nil(t, request.URL.User, "credentials must not be in the request url")
	return username == TestUsername && password == TestPassword
}
