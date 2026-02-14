package models_test

import (
	"strings"
	"testing"

	"github.com/monetr/monetr/server/models"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		id := NewID[Account]()
		assert.NotEmpty(t, id, "generated ID must not be empty")
		assert.True(t, strings.HasPrefix(id.String(), string(AccountIDKind)), "must have the account prefix for this ID")
		assert.False(t, id.IsZero(), "ID should not be zero")

		{ // Parse it correctly
			parsed, err := ParseID[Account](id.String())
			assert.NoError(t, err, "should be able to parse ID as an account ID")
			assert.Equal(t, id, parsed, "parsed ID should match the original ID")
		}

		{ // Parse it incorrectly
			parsed, err := ParseID[User](id.String())
			assert.Error(t, err, "should return an error parsing as the wrong kind of ID")
			assert.True(t, parsed.IsZero(), "bad parsed ID should be zero")
		}
	})
}

func TestParseID(t *testing.T) {
	t.Run("parse bank account ID", func(t *testing.T) {
		input := "bac_123abc"
		output, err := ParseID[BankAccount](input)
		assert.NoError(t, err, "must be able to parse bank account ID")
		assert.EqualValues(t, input, output, "resulting value should match")
		assert.False(t, output.IsZero(), "output should be zero with an error")
	})

	t.Run("bad ID", func(t *testing.T) {
		input := "123abc"
		output, err := ParseID[BankAccount](input)
		assert.EqualError(t, err, "failed to parse ID for models.BankAccount, expected prefix: bac ID: 123abc")
		assert.Empty(t, output, "output should be empty when there is an error")
		assert.True(t, output.IsZero(), "output should be zero with an error")
	})

	t.Run("mismatched data type", func(t *testing.T) {
		input := "user_abc123"
		output, err := ParseID[Account](input)
		assert.EqualError(t, err, "failed to parse ID for models.Account, expected prefix: acct ID: user_abc123")
		assert.Empty(t, output, "output should be empty when there is an error")
		assert.True(t, output.IsZero(), "output should be zero with an error")
	})
}

func TestIDWithoutPrefix(t *testing.T) {
	t.Run("will trim prefix properly", func(t *testing.T) {
		id := models.NewID[models.Account]()
		prefix := (models.Account{}).IdentityPrefix()
		assert.NotContains(t, id.WithoutPrefix(), prefix, "should not contain the prefix")
	})
}
