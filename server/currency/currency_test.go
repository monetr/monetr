package currency_test

import (
	"io"
	"testing"

	"github.com/monetr/monetr/server/currency"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestParseFriendlyToAmount(t *testing.T) {
	t.Run("USD", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1234.99", "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123499, result, "should return an exact int64")
	})

	t.Run("USD weird", func(t *testing.T) {
		// This test looks weird but this particular number parsed by big float and
		// then multiplied by 100 then converted back into a regular integer results
		// in a rounding error. Floating point numbers are the dumbest fucking thing
		// ever.
		result, err := currency.ParseFriendlyToAmount("4315.26", "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 431526, result, "should return an exact int64")
	})

	t.Run("JPY", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1239", "JPY")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 1239, result, "should return an exact int64")
	})

	t.Run("JPY truncation", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1239.99", "JPY")
		assert.EqualError(t, err, "invalid input for currency provided, cannot have more than [0] fractional digits, input: [1239.99], result: [1239.99]")
		assert.EqualValues(t, 0, result, "should return an exact int64")
	})

	t.Run("EUR", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1239.99", "EUR")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123999, result, "should return an exact int64")
	})

	t.Run("invalid currency", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("1239.99", "???")
		assert.EqualError(t, err, "failed to get currency information [???]: currency not supported")
		assert.Zero(t, result, "should return an exact int64")
	})

	t.Run("huge number USD", func(t *testing.T) {
		t.Skip("This test is broken until I can implement something better")
		result, err := currency.ParseFriendlyToAmount("23456789123456789.99", "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, int64(2345678912345678999), result, "should return an exact int64")
	})

	t.Run("overflow USD", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("123456789123456789123456789.99", "USD")
		assert.EqualError(t, err, "overflow, result is larger than a 64-bit integer: [1.234567891e+28]")
		assert.EqualValues(t, 0, result, "should return an exact int64")
	})

	t.Run("empty", func(t *testing.T) {
		result, err := currency.ParseFriendlyToAmount("", "USD")
		assert.EqualError(t, err, "failed to convert string amount to big float: EOF")
		assert.Equal(t, io.EOF, errors.Cause(err), "should be caused by an EOF error")
		assert.EqualValues(t, 0, result, "should return an exact int64")
	})
}

func TestParseFloatToAmount(t *testing.T) {
	t.Run("simple float to amount", func(t *testing.T) {
		// From https://github.com/monetr/monetr/issues/2594 to make sure there is
		// no regression going forward.
		input := 575.67
		result, err := currency.ParseFloatToAmount(input, "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 57567, result, "should")
	})

	t.Run("negative float", func(t *testing.T) {
		result, err := currency.ParseFloatToAmount(-12.34, "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, -1234, result, "should return an exact int64")
	})

	t.Run("float32", func(t *testing.T) {
		result, err := currency.ParseFloatToAmount(float32(0.1), "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 10, result, "should return an exact int64")
	})

	t.Run("big float without an exponent", func(t *testing.T) {
		// fmt.Sprint would have given us 1e+07 here
		result, err := currency.ParseFloatToAmount(10000000.0, "USD")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 1000000000, result, "should return an exact int64")
	})

	t.Run("JPY float", func(t *testing.T) {
		result, err := currency.ParseFloatToAmount(1234.0, "JPY")
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 1234, result, "should return an exact int64")
	})

	t.Run("unsupported currency", func(t *testing.T) {
		result, err := currency.ParseFloatToAmount(1.0, "DEM")
		assert.EqualError(t, err, "failed to get currency information [DEM]: currency not supported")
		assert.Zero(t, result)
	})
}

func TestParseCurrency(t *testing.T) {
	mustGetCurrency := func(t *testing.T, locale, code string) currency.Currency {
		result, err := currency.GetCurrency(locale, code)
		assert.NoError(t, err, "should be able to get the currency")
		return result
	}

	t.Run("EUR", func(t *testing.T) {
		result, err := currency.ParseCurrency("1239.99", mustGetCurrency(t, "en", "EUR"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123999, result, "should return an exact int64")
	})

	t.Run("EUR whole number", func(t *testing.T) {
		result, err := currency.ParseCurrency("1239", mustGetCurrency(t, "en", "EUR"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123900, result, "should return an exact int64")
	})

	t.Run("huge number USD", func(t *testing.T) {
		// Unlike the old implementation, this implementation can handle huge
		// numbers without rounding issues.
		result, err := currency.ParseCurrency("23456789123456789.99", mustGetCurrency(t, "en", "USD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, int64(2345678912345678999), result, "should return an exact int64")
	})

	t.Run("USD with grouping", func(t *testing.T) {
		// This used to eat the digit after the comma
		result, err := currency.ParseCurrency("1,234.56", mustGetCurrency(t, "en", "USD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123456, result, "should return an exact int64")
	})

	t.Run("negative USD with grouping", func(t *testing.T) {
		result, err := currency.ParseCurrency("-1,234,567.89", mustGetCurrency(t, "en", "USD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, -123456789, result, "should return an exact int64")
	})

	t.Run("EUR in german", func(t *testing.T) {
		// German flips the separators around
		result, err := currency.ParseCurrency("1.234,56", mustGetCurrency(t, "de", "EUR"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123456, result, "should return an exact int64")
	})

	t.Run("EUR in french", func(t *testing.T) {
		// French groups with a narrow no-break space which is more than one byte
		result, err := currency.ParseCurrency("1\u202f234,56", mustGetCurrency(t, "fr", "EUR"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123456, result, "should return an exact int64")
	})

	t.Run("CHF in swiss german", func(t *testing.T) {
		result, err := currency.ParseCurrency("1'234.56", mustGetCurrency(t, "de-CH", "CHF"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123456, result, "should return an exact int64")
	})

	t.Run("CVE in cape verde", func(t *testing.T) {
		// Escudos use $ as the decimal separator
		result, err := currency.ParseCurrency("1$50", mustGetCurrency(t, "pt-CV", "CVE"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 150, result, "should return an exact int64")
	})

	t.Run("JPY has no decimal places", func(t *testing.T) {
		result, err := currency.ParseCurrency("1,234", mustGetCurrency(t, "ja", "JPY"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 1234, result, "should return an exact int64")
	})

	t.Run("BHD has three decimal places", func(t *testing.T) {
		result, err := currency.ParseCurrency("1.234", mustGetCurrency(t, "en", "BHD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 1234, result, "should return an exact int64")
	})

	t.Run("blank separators use the defaults", func(t *testing.T) {
		result, err := currency.ParseCurrency("1,234.56", currency.Currency{
			Code:             "USD",
			FractionalDigits: 2,
		})
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123456, result, "should return an exact int64")
	})

	t.Run("negative", func(t *testing.T) {
		// This used to spin forever
		result, err := currency.ParseCurrency("-5.00", mustGetCurrency(t, "en", "USD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, -500, result, "should return an exact int64")
	})

	t.Run("negative in parentheses", func(t *testing.T) {
		result, err := currency.ParseCurrency("(5.00)", mustGetCurrency(t, "en", "USD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, -500, result, "should return an exact int64")
	})

	t.Run("locale minus sign", func(t *testing.T) {
		// Finnish uses a real minus sign, but a plain - should still work too
		fi := mustGetCurrency(t, "fi", "EUR")
		result, err := currency.ParseCurrency("\u221212,50", fi)
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, -1250, result, "should return an exact int64")

		result, err = currency.ParseCurrency("-12,50", fi)
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, -1250, result, "should return an exact int64")
	})

	t.Run("locale minus sign with left-to-right mark", func(t *testing.T) {
		// Arabic has a left-to-right mark in front of the minus sign, it should
		// work with the mark, without it, and with just a plain -
		ar := mustGetCurrency(t, "ar-EG", "EGP")
		for _, input := range []string{"\u200e-5.00", "-5.00"} {
			result, err := currency.ParseCurrency(input, ar)
			assert.NoError(t, err, "should not return an error for %q", input)
			assert.EqualValues(t, -500, result, "should return an exact int64 for %q", input)
		}

		// Persian is both, a left-to-right mark and a real minus sign
		fa := mustGetCurrency(t, "fa", "IRR")
		for _, input := range []string{"\u200e\u22125", "\u22125", "-5"} {
			result, err := currency.ParseCurrency(input, fa)
			assert.NoError(t, err, "should not return an error for %q", input)
			assert.EqualValues(t, -5, result, "should return an exact int64 for %q", input)
		}
	})

	t.Run("minus sign from another locale is not accepted", func(t *testing.T) {
		// en only uses a plain -, so the finnish minus sign is garbage here
		result, err := currency.ParseCurrency("\u22125.00", mustGetCurrency(t, "en", "USD"))
		assert.Error(t, err, "should not parse a minus sign en doesn't use")
		assert.Zero(t, result)
	})

	t.Run("trailing negative", func(t *testing.T) {
		usd := mustGetCurrency(t, "en", "USD")
		for _, input := range []string{"5.00-", "5.00 -"} {
			result, err := currency.ParseCurrency(input, usd)
			assert.NoError(t, err, "should not return an error for %q", input)
			assert.EqualValues(t, -500, result, "should return an exact int64 for %q", input)
		}
	})

	t.Run("space after the negative", func(t *testing.T) {
		usd := mustGetCurrency(t, "en", "USD")
		for _, input := range []string{"- 5.00", "( 5.00 )"} {
			result, err := currency.ParseCurrency(input, usd)
			assert.NoError(t, err, "should not return an error for %q", input)
			assert.EqualValues(t, -500, result, "should return an exact int64 for %q", input)
		}
	})

	t.Run("extra decimal places are truncated", func(t *testing.T) {
		result, err := currency.ParseCurrency("1.2349", mustGetCurrency(t, "en", "USD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 123, result, "should return an exact int64")
	})

	t.Run("extra decimal places round up", func(t *testing.T) {
		result, err := currency.ParseCurrency("1.006", mustGetCurrency(t, "en", "USD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, 101, result, "should return an exact int64")

		result, err = currency.ParseCurrency("-1.006", mustGetCurrency(t, "en", "USD"))
		assert.NoError(t, err, "should not return an error")
		assert.EqualValues(t, -101, result, "should return an exact int64")
	})

	t.Run("garbage", func(t *testing.T) {
		// This also used to spin forever instead of returning an error
		result, err := currency.ParseCurrency("12abc", mustGetCurrency(t, "en", "USD"))
		assert.EqualError(t, err, "failed to parse currency 12abc - USD, unexpected character a")
		assert.Zero(t, result)
	})
}
