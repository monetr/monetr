package teller_test

import (
	"testing"

	"github.com/monetr/monetr/server/teller"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_GetDescription(t *testing.T) {
	t.Run("multiline description", func(t *testing.T) {
		transaction := teller.Transaction{
			// Some banks have their transactions formatted like this.
			Description: "POS Debit - Visa Check Card 1234 -\n\t\t\t\t\tAMZN Mktp US Seattle WAUS",
		}

		description := transaction.GetDescription()
		assert.Equal(t,
			"POS Debit - Visa Check Card 1234 - AMZN Mktp US Seattle WAUS",
			description,
			"description should be sanitized",
		)
	})
}

func TestTransaction_GetAmount(t *testing.T) {
	t.Run("debit", func(t *testing.T) {
		transaction := teller.Transaction{
			Amount: "-10.12",
		}
		amount, err := transaction.GetAmount()
		assert.NoError(t, err, "should not have an error parsing amount")
		assert.EqualValues(t, 1012, amount, "amount should now the inverted cents")
	})
}
