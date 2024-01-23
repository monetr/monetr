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

	bankAccount := GivenIHaveABankAccount(t, clock, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)
	assert.NotZero(t, bankAccount.BankAccountId, "bank account must have been created")
}
