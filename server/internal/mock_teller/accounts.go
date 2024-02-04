package mock_teller

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/mock_http_helper"
	"github.com/monetr/monetr/server/teller"
	"github.com/stretchr/testify/require"
)

func BankAccountFixture(t *testing.T) teller.Account {
	accountNumber := gofakeit.AchAccount()
	require.NotEmpty(t, accountNumber, "account number cannot be empty")

	accountType := teller.AccountType(gofakeit.RandomString([]string{
		string(teller.AccountTypeDepository),
		string(teller.AccountTypeDepository),
		string(teller.AccountTypeDepository),
		string(teller.AccountTypeCredit),
	}))

	accountSubTypes := map[teller.AccountType]teller.AccountSubType{
		teller.AccountTypeDepository: teller.AccountSubType(gofakeit.RandomString([]string{
			string(teller.AccountSubTypeChecking),
			string(teller.AccountSubTypeSavings),
			string(teller.AccountSubTypeMoneyMarket),
			string(teller.AccountSubTypeCertificateOfDeposit),
		})),
		teller.AccountTypeCredit: teller.AccountSubTypeCreditCard,
	}
	accountSubType := accountSubTypes[accountType]

	mask := accountNumber[len(accountNumber)-4:]

	currencyCode := "USD"

	city := gofakeit.City()

	return teller.Account{
		Id:           gofakeit.Generate("acc_???????????????????"),
		Currency:     currencyCode,
		EnrollmentId: gofakeit.Generate("enr_???????????????????"),
		Institution: struct {
			Id   string "json:\"id\""
			Name string "json:\"name\""
		}{
			Id:   strings.Join(strings.Split(strings.ToLower(city), " "), "_"),
			Name: fmt.Sprintf("Bank of %s", city),
		},
		Mask:    mask,
		Links:   map[string]string{},
		Name:    "My Checking",
		Type:    accountType,
		SubType: accountSubType,
		Status:  "open",
	}
}

func MockGetAccounts(t *testing.T, accounts []teller.Account) {
	mock_http_helper.NewHttpMockJsonResponder(
		t,
		"GET", Path(t, "/accounts"),
		func(t *testing.T, request *http.Request) (interface{}, int) {
			ValidateTellerAuthentication(t, request, RequireAccessToken)
			return accounts, http.StatusOK
		},
		nil,
	)
}
