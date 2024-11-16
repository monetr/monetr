package models_test

import (
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestParseBankAccountType(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		values := []string{
			"depository",
			"credit",
			"loan",
			"investment",
		}

		for _, value := range values {
			assert.NotEqual(t,
				models.OtherBankAccountType, models.ParseBankAccountType(value),
				"Must not return an other account type",
			)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		values := []string{
			"something?",
		}

		for _, value := range values {
			assert.Equal(t,
				models.OtherBankAccountType, models.ParseBankAccountType(value),
				"Must return an other account type",
			)
		}
	})
}

func TestParseBankAccountSubType(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		values := []string{
			"checking",
			"savings",
			"auto",
			"credit card",
		}

		for _, value := range values {
			assert.NotEqual(t,
				models.OtherBankAccountSubType, models.ParseBankAccountSubType(value),
				"Must not return an other sub type",
			)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		values := []string{
			"something?",
			"mortgage", // Not yet supported
		}

		for _, value := range values {
			assert.Equal(t,
				models.OtherBankAccountSubType, models.ParseBankAccountSubType(value),
				"Must return an other sub type",
			)
		}
	})
}

func TestParseBankAccountStatus(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		values := []string{
			"active",
			"inactive",
		}

		for _, value := range values {
			assert.NotEqual(t,
				models.UnknownBankAccountStatus, models.ParseBankAccountStatus(value),
				"Must not return an unknown status",
			)
		}
	})

	t.Run("accepts bank account status types", func(t *testing.T) {
		values := []models.BankAccountStatus{
			"active",
			"inactive",
		}

		for _, value := range values {
			assert.NotEqual(t,
				models.UnknownBankAccountStatus, models.ParseBankAccountStatus(value),
				"Must not return an unknown status",
			)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		assert.Equal(t,
			models.UnknownBankAccountStatus, models.ParseBankAccountStatus("something"),
			"Must return an unknown status for an invalid input",
		)
	})
}
