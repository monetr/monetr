package fixtures

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/models"
	"testing"
)

func BankAccountFixture(t *testing.T) *models.BankAccount {
	return &models.BankAccount{
		PlaidAccountId:    gofakeit.UUID(),
		AvailableBalance:  100000,
		CurrentBalance:    98500,
		Mask:              "0123",
		Name:              "Personal Checking",
		PlaidName:         "Checking",
		PlaidOfficialName: "Checking",
		Type:              "depository",
		SubType:           "checking",
	}
}
