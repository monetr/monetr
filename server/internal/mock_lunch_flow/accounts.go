package mock_lunch_flow

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// TestInvalidAPIToken is always considered invalid if used in an API request
	// to the mock endpoints.
	TestInvalidAPIToken = "test-token-that-is-always-invalid"
)

func Path(t *testing.T, relative string) string {
	require.NotEmpty(t, relative, "relative url cannot be empty")
	parsed, err := url.Parse(lunch_flow.DefaultBaseURL)
	require.NoError(t, err, "must be able to parse lunch flow's default base URL")
	parsed.Path = relative
	return parsed.String()
}

const (
	RequireAccessToken      = true
	DoNotRequireAccessToken = false
)

func ValidateLunchFlowAuthentication(
	t *testing.T,
	request *http.Request,
	requireAccessToken bool,
) (accessToken string) {
	original := request.Header.Get("x-api-key")
	authorization := strings.TrimSpace(original)
	// Only perform assertions if they are required by the caller. If they are not
	// then simply parse the authorization and return it to the caller. This way
	// things like unauthorized errors can be implemented by the mock handler!
	if requireAccessToken {
		assert.NotEqual(t, "", authorization, "Token cannot be empty!")
		assert.Equal(t, authorization, original, "Bearer token format does not match the expected format")
	}
	return authorization
}

func LunchFlowHeaders(
	t *testing.T,
	request *http.Request,
	response any,
	status int,
) map[string][]string {
	return map[string][]string{
		"Content-Type": {
			"application/json",
		},
		"X-Vercel-Id": {
			// In the real world this would look like this:
			// `cle1::lhr1::lr685-1768772030274-6bfe469de05f`
			// But for tests its fine to use anything here thats unique.
			gofakeit.UUID(),
		},
	}
}

func MockFetchAccounts(t *testing.T, accounts []lunch_flow.Account) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"GET", Path(t, "/api/v1/accounts"),
		func(t *testing.T, request *http.Request) (any, int) {
			if token := ValidateLunchFlowAuthentication(
				t,
				request,
				RequireAccessToken,
			); token == TestInvalidAPIToken {
				return map[string]any{
					"error":   "Forbidden",
					"message": "Invalid API key.",
				}, http.StatusForbidden
			}

			return map[string]any{
				"accounts": accounts,
				"total":    len(accounts),
			}, http.StatusOK
		},
		LunchFlowHeaders,
	)
}

func MockFetchAccountsError(t *testing.T) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"GET", Path(t, "/api/v1/accounts"),
		func(t *testing.T, request *http.Request) (any, int) {
			return map[string]any{
				"error":   "Forbidden",
				"message": "Invalid API key.",
			}, http.StatusForbidden
		},
		LunchFlowHeaders,
	)
}

func MockFetchBalance(
	t *testing.T,
	accountId lunch_flow.AccountId,
	balance lunch_flow.Balance,
) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"GET", Path(t, fmt.Sprintf("/api/v1/accounts/%s/balance", accountId)),
		func(t *testing.T, request *http.Request) (any, int) {
			if token := ValidateLunchFlowAuthentication(
				t,
				request,
				RequireAccessToken,
			); token == TestInvalidAPIToken {
				return map[string]any{
					"error":   "Forbidden",
					"message": "Invalid API key.",
				}, http.StatusForbidden
			}

			return map[string]any{
				"balance": balance,
			}, http.StatusOK
		},
		LunchFlowHeaders,
	)
}
