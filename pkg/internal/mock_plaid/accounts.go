package mock_plaid

import (
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/rest-api/pkg/internal/mock_http_helper"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/monetr/rest-api/pkg/internal/testutils"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
)

func BankAccountFixture(t *testing.T) plaid.AccountBase {
	accountNumber := gofakeit.AchAccount()
	require.NotEmpty(t, accountNumber, "account number cannot be empty")

	accountType := gofakeit.RandomString([]string{
		string(plaid.ACCOUNTTYPE_DEPOSITORY),
		string(plaid.ACCOUNTTYPE_CREDIT),
		string(plaid.ACCOUNTTYPE_INVESTMENT),
		string(plaid.ACCOUNTTYPE_LOAN),
	})

	var accountSubType plaid.AccountSubtype
	switch plaid.AccountType(accountType) {
	case plaid.ACCOUNTTYPE_DEPOSITORY:
		accountSubType = plaid.AccountSubtype(gofakeit.RandomString([]string{
			string(plaid.ACCOUNTSUBTYPE_CHECKING),
			string(plaid.ACCOUNTSUBTYPE_SAVINGS),
			string(plaid.ACCOUNTSUBTYPE_PAYPAL),
		}))
	case plaid.ACCOUNTTYPE_CREDIT:
		accountSubType = plaid.AccountSubtype(gofakeit.RandomString([]string{
			string(plaid.ACCOUNTSUBTYPE_CREDIT_CARD),
			string(plaid.ACCOUNTSUBTYPE_PAYPAL),
		}))
	case plaid.ACCOUNTTYPE_INVESTMENT:
		accountSubType = plaid.AccountSubtype(gofakeit.RandomString([]string{
			string(plaid.ACCOUNTSUBTYPE_IRA),
			string(plaid.ACCOUNTSUBTYPE_ROTH),
		}))
	case plaid.ACCOUNTTYPE_LOAN:
		accountSubType = plaid.AccountSubtype(gofakeit.RandomString([]string{
			string(plaid.ACCOUNTSUBTYPE_AUTO),
			string(plaid.ACCOUNTSUBTYPE_HOME),
		}))
	}

	mask := accountNumber[len(accountNumber)-4:]

	currencyCode := "USD"

	current := gofakeit.Float32Range(100, 500)
	available := gofakeit.Float32Range(current-10, current)
	limit := gofakeit.Float32Range(current, current+100)

	return plaid.AccountBase{
		AccountId: gofakeit.Generate("????????????????"),
		Balances: plaid.AccountBalance{
			Available:              *plaid.NewNullableFloat32(myownsanity.Float32P(available)),
			Current:                *plaid.NewNullableFloat32(myownsanity.Float32P(current)),
			Limit:                  *plaid.NewNullableFloat32(myownsanity.Float32P(limit)),
			IsoCurrencyCode:        *plaid.NewNullableString(myownsanity.StringP(currencyCode)),
			UnofficialCurrencyCode: *plaid.NewNullableString(myownsanity.StringP(currencyCode)),
		},
		Mask:         *plaid.NewNullableString(myownsanity.StringP(mask)),
		Name:         fmt.Sprintf("Personal Account - %s", mask),
		OfficialName: *plaid.NewNullableString(myownsanity.StringP(fmt.Sprintf("%s - %s", strings.ToUpper(accountType), mask))),
		Type:         plaid.AccountType(accountType),
		Subtype:      *plaid.NewNullableAccountSubtype(&accountSubType),
	}
}

func MockGetAccountsExtended(t *testing.T, plaidData *testutils.MockPlaidData) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/accounts/get"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			accessToken := ValidatePlaidAuthentication(t, request, RequireAccessToken)
			var getAccountsRequest struct {
				Options     struct {
					 AccountIds []string `json:"account_ids"`
				} `json:"options"`
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&getAccountsRequest), "must decode request")

			accounts, ok := plaidData.BankAccounts[accessToken]
			require.True(t, ok, "invalid access token mocking not implemented")

			response := plaid.AccountsGetResponse{
				RequestId: gofakeit.UUID(),
				Accounts:  make([]plaid.AccountBase, 0),
				Item:      plaid.Item{}, // Not yet populating this.
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

func MockGetAccounts(t *testing.T, accounts []plaid.AccountBase) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/accounts/get"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
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
				"accounts": accounts,
			}, http.StatusOK
		},
		PlaidHeaders,
	)
}

func MockGetAccountsError(t *testing.T, plaidError plaid.Error) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"POST", Path(t, "/accounts/get"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
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
			if s := plaidError.Status.Get(); s != nil {
				status = int(*s)
			} else {
				status = http.StatusInternalServerError
			}

			return plaidError, status
		},
		PlaidHeaders,
	)
}
