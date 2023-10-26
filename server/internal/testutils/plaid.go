package testutils

import (
	"github.com/monetr/monetr/server/models"
	"github.com/plaid/plaid-go/v14/plaid"
)

type MockPlaidData struct {
	PlaidTokens  map[string]models.PlaidToken
	PlaidLinks   map[string]models.PlaidLink
	BankAccounts map[string]map[string]plaid.AccountBase
}
