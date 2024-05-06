package models_test

import (
	"testing"

	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestParseID(t *testing.T) {
	t.Run("parse bank account ID", func(t *testing.T) {
		input := "bac_123abc"
		output, err := ParseID[BankAccount](input)
		assert.NoError(t, err, "must be able to parse bank acccount ID")
		assert.EqualValues(t, input, output, "resulting value should match")
		assert.False(t, output.IsZero(), "output should be zero with an error")
	})

	t.Run("bad ID", func(t *testing.T) {
		input := "123abc"
		output, err := ParseID[BankAccount](input)
		assert.EqualError(t, err, "failed to parse ID for BankAccount, expected prefix: bac_ ID: 123abc")
		assert.Empty(t, output, "output should be empty when there is an error")
		assert.True(t, output.IsZero(), "output should be zero with an error")
	})
}
