package fixtures

import (
	"testing"

	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGivenIHaveATransaction(t *testing.T) {
	user, _ := GivenIHaveABasicAccount(t)
	link := GivenIHaveAPlaidLink(t, user)
	bankAccount := GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

	transaction := GivenIHaveATransaction(t, bankAccount)
	assert.NotZero(t, transaction.TransactionId, "transaction must have been created")
	assert.NotNil(t, transaction.Account, "account must be included on the transaction")
	assert.NotNil(t, transaction.BankAccount, "bank account must be included on the transaction")
	assert.Greater(t, transaction.Amount, int64(0), "amount must be greater than 0")
}
