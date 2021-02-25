package testutils

import "github.com/plaid/plaid-go/plaid"

type MockPlaidData struct {
	BankAccounts map[string]map[string]plaid.Account
}
