package testutils

import (
	"github.com/monetr/monetr/pkg/models"
	"github.com/plaid/plaid-go/plaid"
)

type MockPlaidData struct {
	PlaidTokens  map[string]models.PlaidToken
	PlaidLinks   map[string]models.PlaidLink
	BankAccounts map[string]map[string]plaid.AccountBase
}
