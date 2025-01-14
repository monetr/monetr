package currency_test

import (
	"math"
	"testing"

	"github.com/monetr/monetr/server/currency"
	"github.com/stretchr/testify/assert"
)

func TestParseFriendlyToAmount(t *testing.T) {
	t.Run("USD", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1234.99", "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123499, result, "should return an exact int64")
	})

	t.Run("JPY", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1239", "JPY")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 1239, result, "should return an exact int64")
	})

	t.Run("JPY truncation", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1239.99", "JPY")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 1239, result, "should return an exact int64")
	})

	t.Run("EUR", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1239.99", "EUR")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123999, result, "should return an exact int64")
	})

	t.Run("invalid currency", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1239.99", "???")
		assert.EqualError(t, err, "failed to parse currency amount: currency not supported")
		assert.Zero(t, result, "should return an exact int64")
	})

	t.Run("huge number USD", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("23456789123456789.99", "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, int64(2345678912345678999), result, "should return an exact int64")
	})

	t.Run("overflow USD", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("123456789123456789123456789.99", "USD")
		assert.EqualError(t, err, "failed to parse currency amount accurately: Below")
		assert.EqualValues(t, math.MaxInt64, result, "should return an exact int64")
	})
}
