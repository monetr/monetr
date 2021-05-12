package testutils

import (
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/plaid/plaid-go/plaid"
)

type MockPlaidData struct {
	PlaidLinks   map[string]models.PlaidLink
	BankAccounts map[string]map[string]plaid.Account
}
