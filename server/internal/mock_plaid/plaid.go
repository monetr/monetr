package mock_plaid

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

func PlaidHeaders(t *testing.T, request *http.Request, response any, status int) map[string][]string {
	return map[string][]string{
		"X-Request-Id": {
			gofakeit.UUID(),
		},
	}
}
