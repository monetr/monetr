package fixtures

import (
	"testing"
	"time"

	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestGivenIHaveAFundingSchedule(t *testing.T) {
	user, _ := GivenIHaveABasicAccount(t)
	link := GivenIHaveAPlaidLink(t, user)
	bankAccount := GivenIHaveABankAccount(t, &link, models.DepositoryBankAccountType, models.CheckingBankAccountSubType)

	fundingSchedule := GivenIHaveAFundingSchedule(t, &bankAccount, "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1", false)
	assert.NotZero(t, fundingSchedule.FundingScheduleId, "newly created funding schedule must have an Id")
	assert.NotNil(t, fundingSchedule.BankAccount, "should have the bank account field set")
	assert.NotNil(t, fundingSchedule.Account, "should have the account field set")
	assert.Equal(t, bankAccount.BankAccountId, fundingSchedule.BankAccountId, "should have the expected bank account Id")
	assert.NotEmpty(t, fundingSchedule.Name, "should have a funding schedule name")
	assert.NotEmpty(t, fundingSchedule.NextOccurrence, "next occurrence should be set")
	assert.Greater(t, fundingSchedule.NextOccurrence, time.Now(), "next occurrence should be in the future")
}
