package fixtures

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestGivenIHaveABankAccount(t *testing.T) {
	clock := clock.NewMock()
	user, _ := GivenIHaveABasicAccount(t, clock)
	link := GivenIHaveAPlaidLink(t, clock, user)

	assert.Len(t, link.BankAccounts, 0, "there should be no bank accounts initially")

	bankAccount := GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
	assert.NotZero(t, bankAccount.BankAccountId, "bank account must have been created")
	assert.Len(t, link.BankAccounts, 1, "there should now be a single bank account")
}
