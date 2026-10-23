package mock_plaid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/monetr/monetr/server/internal/testutils"
	plaid "github.com/plaid/plaid-go/v49"
	"github.com/stretchr/testify/require"
)

func BankAccountFixture(t *testing.T) *plaid.AccountBase {
	accountNumber := gofakeit.AchAccount()
	require.NotEmpty(t, accountNumber, "account number cannot be empty")

	accountType := gofakeit.RandomString([]string{
		string(plaid.AccountTypeDepository),
		string(plaid.AccountTypeCredit),
		string(plaid.AccountTypeInvestment),
		string(plaid.AccountTypeLoan),
	})

	accountSubTypes := map[plaid.AccountType]plaid.AccountSubtype{
		plaid.AccountTypeDepository: plaid.AccountSubtype(gofakeit.RandomString([]string{
			string(plaid.AccountSubtypeChecking),
			string(plaid.AccountSubtypeSavings),
			string(plaid.AccountSubtypePaypal),
		})),
		plaid.AccountTypeCredit: plaid.AccountSubtype(gofakeit.RandomString([]string{
			string(plaid.AccountSubtypeCreditCard),
			string(plaid.AccountSubtypePaypal),
		})),
		plaid.AccountTypeInvestment: plaid.AccountSubtype(gofakeit.RandomString([]string{
			string(plaid.AccountSubtypeIra),
			string(plaid.AccountSubtypeRoth),
		})),
		plaid.AccountTypeLoan: plaid.AccountSubtype(gofakeit.RandomString([]string{
			string(plaid.AccountSubtypeAuto),
			string(plaid.AccountSubtypeMortgage),
		})),
	}
	accountSubType := accountSubTypes[plaid.AccountType(accountType)]

	mask := accountNumber[len(accountNumber)-4:]

	currencyCode := "USD"

	current := gofakeit.Float64Range(100, 500)
	available := gofakeit.Float64Range(current-10, current)
	limit := gofakeit.Float64Range(current, current+100)

	return &plaid.AccountBase{
		AccountID: gofakeit.Lexify("????????????????"),
		Balances: &plaid.AccountBalance{
			Available:              new(available),
			Current:                new(current),
			Limit:                  new(limit),
			IsoCurrencyCode:        new(currencyCode),
			UnofficialCurrencyCode: new(currencyCode),
		},
		Mask:         new(mask),
		Name:         fmt.Sprintf("Personal Account - %s", mask),
		OfficialName: new(fmt.Sprintf("%s - %s", strings.ToUpper(accountType), mask)),
		Type:         plaid.AccountType(accountType),
		Subtype:      &accountSubType,
	}
}

func MockGetAccountsExtended(t *testing.T, plaidData *testutils.MockPlaidData) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/accounts/get"),
		func(t *testing.T, request *http.Request) (any, int) {
			accessToken := ValidatePlaidAuthentication(t, request, RequireAccessToken)
			var getAccountsRequest struct {
				Options struct {
					AccountIds []string `json:"account_ids"`
				} `json:"options"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&getAccountsRequest), "must decode request")

			accounts, ok := plaidData.BankAccounts[accessToken]
			require.True(t, ok, "invalid access token mocking not implemented")

			response := &plaid.AccountsGetResponse{
				RequestID: gofakeit.UUID(),
				Accounts:  make([]*plaid.AccountBase, 0),
				Item:      &plaid.Item{}, // Not yet populating this.
			}
			for _, accountId := range getAccountsRequest.Options.AccountIds {
				account, ok := accounts[accountId]
				if !ok {
					panic("bad account id handling not yet implemented")
				}

				response.Accounts = append(response.Accounts, account)
			}

			return response, http.StatusOK
		},
		PlaidHeaders,
	)
}

func MockGetAccounts(t *testing.T, accounts []*plaid.AccountBase) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/accounts/get"),
		func(t *testing.T, request *http.Request) (any, int) {
			var getAccountsRequest struct {
				ClientId    string `json:"client_id"`
				Secret      string `json:"secret"`
				AccessToken string `json:"access_token"`
				Options     struct {
					AccountIds []string `json:"account_ids"`
				} `json:"options"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&getAccountsRequest), "must decode request")

			return map[string]any{
				"accounts": accounts,
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}

func MockGetAccountsError(t *testing.T, plaidError plaid.PlaidError) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/accounts/get"),
		func(t *testing.T, request *http.Request) (any, int) {
			var getAccountsRequest struct {
				ClientId    string `json:"client_id"`
				Secret      string `json:"secret"`
				AccessToken string `json:"access_token"`
				Options     struct {
					AccountIds []string `json:"account_ids"`
				} `json:"options"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&getAccountsRequest), "must decode request")

			var status int
			if s := plaidError.Status; s != nil {
				status = *s
			} else {
				status = http.StatusInternalServerError
			}

			return &plaidError, status
		},
		PlaidHeaders,
	)
}
