package mock_lunch_flow

import (
	"net/http"
	"testing"

	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
)

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
