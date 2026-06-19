package mock_plaid

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
)

func PlaidHeaders(_ *testing.T, _ *http.Request, _ any, _ int) map[string][]string {
	return map[string][]string{
		"X-Request-Id": {
			gofakeit.UUID(),
		},
	}
}
