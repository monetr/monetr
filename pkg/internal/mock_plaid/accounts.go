package mock_plaid

import (
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/internal/mock_http_helper"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/internal/testutils"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func MockGetAccountsExtended(t *testing.T, plaidData *testutils.MockPlaidData) {
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

		accounts, ok := plaidData.BankAccounts[getAccountsRequest.AccessToken]
		if !ok {
			panic("invalid access token mocking not implemented")
		}

		response := plaid.GetAccountsResponse{
			APIResponse: plaid.APIResponse{
				RequestID: gofakeit.UUID(),
			},
			Accounts: make([]plaid.Account, 0),
			Item:     plaid.Item{}, // Not yet populating this.
		}
		for _, accountId := range getAccountsRequest.Options.AccountIds {
			account, ok := accounts[accountId]
			if !ok {
				panic("bad account id handling not yet implemented")
			}

			response.Accounts = append(response.Accounts, account)
		}

		return response, http.StatusOK
	})
}

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
