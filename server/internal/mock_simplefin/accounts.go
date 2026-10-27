package mock_simplefin

import (
	"net/http"
	"testing"

	"github.com/monetr/monetr/server/datasources/simplefin"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/stretchr/testify/assert"
)

func MockFetchAccounts(t *testing.T, accountSet simplefin.AccountSet) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"GET", Path(t, "/simplefin/accounts"),
		func(t *testing.T, request *http.Request) (any, int) {
			if !ValidateSimpleFinAuthentication(t, request) {
				return "Forbidden", http.StatusForbidden
			}
			assert.Equal(t, "2", request.URL.Query().Get("version"), "must request version 2 of the protocol")

			return accountSet, http.StatusOK
		},
		nil,
	)
}

// MockFetchAccountsError will always respond with the provided status code, use
// [http.StatusPaymentRequired] or [http.StatusForbidden] to simulate the errors
// SimpleFIN documents.
func MockFetchAccountsError(t *testing.T, status int) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"GET", Path(t, "/simplefin/accounts"),
		func(_ *testing.T, _ *http.Request) (any, int) {
			return http.StatusText(status), status
		},
		nil,
	)
}
