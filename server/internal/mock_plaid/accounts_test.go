package mock_plaid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBankAccountFixture(t *testing.T) {
	account := BankAccountFixture(t)
	assert.NotEmpty(t, account, "account must not be empty")
	assert.NotEmpty(t, account.GetAccountID(), "bank account ID must not be empty")

	balances := account.GetBalances()
	assert.NotEmpty(t, balances, "balances must not be empty")
	assert.NotZero(t, *balances.Available, "available must not be zero")
	assert.NotZero(t, *balances.Current, "available must not be zero")
	assert.LessOrEqual(t, *balances.Available, *balances.Current, "available must always be less than or equal to current")
}
