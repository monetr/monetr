package mock_plaid

import (
	"github.com/brianvoe/gofakeit/v6"
	"net/http"
	"testing"
)

func PlaidHeaders(t *testing.T, request *http.Request, response interface{}, status int) map[string][]string {
	return map[string][]string{
		"X-Request-Id": {
			gofakeit.UUID(),
		},
	}
}
