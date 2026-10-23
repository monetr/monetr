package testutils

import (
	"github.com/monetr/monetr/server/models"
	plaid "github.com/plaid/plaid-go/v49"
)

type MockPlaidData struct {
	PlaidTokens  map[string]models.PlaidToken
	PlaidLinks   map[string]models.PlaidLink
	BankAccounts map[string]map[string]*plaid.AccountBase
}
