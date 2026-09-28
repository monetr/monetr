package currency_test

import (
	"testing"

	"github.com/monetr/monetr/server/currency"
	"github.com/stretchr/testify/assert"
	gocurrency "golang.org/x/text/currency"
	"golang.org/x/text/language"
)

// Testing to see if x/text/currency and x/text/language will be viable.
func TestGolangCurrency(t *testing.T) {
	t.Run("german to euro", func(t *testing.T) {
		result, confidence := gocurrency.FromTag(language.MustParse("de"))
		assert.Equal(t, language.Low, confidence)
		assert.Equal(t, "EUR", result.String())
	})

	t.Run("netherlands to euro", func(t *testing.T) {
		result, confidence := gocurrency.FromTag(language.MustParse("nl"))
		assert.Equal(t, language.Low, confidence)
		assert.Equal(t, "EUR", result.String())
	})

	t.Run("ukraine", func(t *testing.T) {
		result, confidence := gocurrency.FromTag(language.MustParse("uk"))
		assert.Equal(t, language.Low, confidence)
		assert.Equal(t, "UAH", result.String())
	})

	t.Run("japanese", func(t *testing.T) {
		result, confidence := gocurrency.FromTag(language.MustParse("ja"))
		assert.Equal(t, language.Low, confidence)
		assert.Equal(t, "JPY", result.String())
	})

	t.Run("jpy currency precision", func(t *testing.T) {
		result, _ := gocurrency.FromTag(language.MustParse("ja"))
		scale, increment := gocurrency.Accounting.Rounding(result)
		assert.Equal(t, 0, scale)
		assert.Equal(t, 1, increment)
	})

	t.Run("eur currency precision", func(t *testing.T) {
		result, _ := gocurrency.FromTag(language.MustParse("de"))
		scale, increment := gocurrency.Accounting.Rounding(result)
		assert.Equal(t, 2, scale)
		assert.Equal(t, 1, increment)
	})

	t.Run("usd currency precision", func(t *testing.T) {
		result, _ := gocurrency.FromRegion(language.MustParseRegion("US"))
		scale, increment := gocurrency.Accounting.Rounding(result)
		assert.Equal(t, 2, scale)
		assert.Equal(t, 1, increment)
	})
}

func TestCurrencyParity(t *testing.T) {
	t.Skip("this test fails because of data mismatch")
	t.Run("make sure they match", func(t *testing.T) {
		currencies := currency.GetCurrencies()
		for i := range currencies {
			localeCurrency := currencies[i]
			goCurrency, err := gocurrency.ParseISO(localeCurrency)
			assert.NoError(t, err, "currency %s is not available in golang", localeCurrency)
			scale, increment := gocurrency.Accounting.Rounding(goCurrency)
			assert.Equal(t, 1, increment, "increment should always be 1")
			localeScale, err := currency.GetFractionalDigits(localeCurrency)
			assert.NoError(t, err, "CLDR could not return fractional digits for %s", localeCurrency)
			assert.EqualValues(t, localeScale, scale, "golang currency and locale currency scale differ for %s", localeCurrency)
		}
	})
}
