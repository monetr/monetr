package mock_plaid

import (
	"encoding/json"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/internal/mock_http_helper"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func MockGetAccounts(t *testing.T, accounts []plaid.Account) {
	mock_http_helper.NewHttpMockJsonResponder(t, "POST", Path(t, "/accounts/get"), func(t *testing.T, request *http.Request) (interface{}, int) {
		var getAccountsRequest struct {
			ClientId    string `json:"client_id"`
			Secret      string `json:"secret"`
			AccessToken string `json:"access_token"`
			Options     struct {
				AccountIds []string `json:"account_ids"`
			} `json:"options"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&getAccountsRequest), "must decode request")

		return map[string]interface{}{
			"accounts": []string{},
		}, http.StatusOK
	})
}
