package mock_lunch_flow

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
)

func MockFetchTransactions(
	t *testing.T,
	accountId lunch_flow.AccountId,
	transactions []lunch_flow.Transaction,
) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"GET", Path(t, fmt.Sprintf("/api/v1/accounts/%s/transactions", accountId)),
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
				"transactions": transactions,
				"total":        len(transactions),
			}, http.StatusOK
		},
		LunchFlowHeaders,
	)
}

func MockFetchTransactionsError(
	t *testing.T,
	accountId lunch_flow.AccountId,
) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"GET", Path(t, fmt.Sprintf("/api/v1/accounts/%s/transactions", accountId)),
		func(t *testing.T, request *http.Request) (any, int) {
			return map[string]any{
				"error":   "Forbidden",
				"message": "Invalid API key.",
			}, http.StatusForbidden
		},
		LunchFlowHeaders,
	)
}
