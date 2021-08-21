package mock_plaid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBankAccountFixture(t *testing.T) {
	account := BankAccountFixture(t)
	assert.NotEmpty(t, account, "account must not be empty")
}
