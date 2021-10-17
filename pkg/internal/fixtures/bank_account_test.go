package fixtures

import (
	"testing"

	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGivenIHaveABankAccount(t *testing.T) {
	user, _ := GivenIHaveABasicAccount(t)
	link := GivenIHaveAPlaidLink(t, user)

	assert.Len(t, link.BankAccounts, 0, "there should be no bank accounts initially")

	bankAccount := GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
	assert.NotZero(t, bankAccount.BankAccountId, "bank account must have been created")
	assert.Len(t, link.BankAccounts, 1, "there should now be a single bank account")
}
