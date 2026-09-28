package currency_test

import (
	"slices"
	"testing"

	"github.com/monetr/monetr/server/currency"
	"github.com/stretchr/testify/assert"
)

func TestParseLocale(t *testing.T) {
	t.Run("bcp 47", func(t *testing.T) {
		tag, err := currency.ParseLocale("en-US")
		assert.NoError(t, err, "should parse a regular locale")
		assert.Equal(t, "en-US", tag.String())
	})

	t.Run("posix", func(t *testing.T) {
		tag, err := currency.ParseLocale("en_US")
		assert.NoError(t, err, "should parse a posix locale")
		assert.Equal(t, "en-US", tag.String())
	})

	t.Run("posix with encoding", func(t *testing.T) {
		tag, err := currency.ParseLocale("de_CH.UTF-8")
		assert.NoError(t, err, "should strip the encoding off")
		assert.Equal(t, "de-CH", tag.String())
	})

	t.Run("posix with modifier", func(t *testing.T) {
		tag, err := currency.ParseLocale("de_DE@euro")
		assert.NoError(t, err, "should strip the modifier off")
		assert.Equal(t, "de-DE", tag.String())
	})

	t.Run("unicode extension", func(t *testing.T) {
		// The browser can give us this from Intl, the extension should be kept but
		// not mistaken for anything else
		tag, err := currency.ParseLocale("en-u-nu-latn")
		assert.NoError(t, err, "should parse a locale with an extension")
		assert.Equal(t, "en-u-nu-latn", tag.String())
	})

	t.Run("blank", func(t *testing.T) {
		_, err := currency.ParseLocale("  ")
		assert.EqualError(t, err, "locale cannot be blank")
	})

	t.Run("garbage", func(t *testing.T) {
		_, err := currency.ParseLocale("!!")
		assert.Error(t, err, "should not parse garbage")
	})
}

func TestGetCurrencyForLocale(t *testing.T) {
	t.Run("with a region", func(t *testing.T) {
		cases := map[string]string{
			"en_US":       "USD",
			"en_US.UTF-8": "USD",
			"de_DE@euro":  "EUR",
			"de_CH":       "CHF",
			"en-GB":       "GBP",
			"zh_TW":       "TWD",
		}
		for locale, expected := range cases {
			code, err := currency.GetCurrencyForLocale(locale)
			assert.NoError(t, err, "should find a currency for %s", locale)
			assert.Equal(t, expected, code, "wrong currency for %s", locale)
		}
	})

	t.Run("language only", func(t *testing.T) {
		cases := map[string]string{
			"en": "USD",
			"ja": "JPY",
			"uk": "UAH",
			"nl": "EUR",
		}
		for locale, expected := range cases {
			code, err := currency.GetCurrencyForLocale(locale)
			assert.NoError(t, err, "should find a currency for %s", locale)
			assert.Equal(t, expected, code, "wrong currency for %s", locale)
		}
	})

	t.Run("script changes the region", func(t *testing.T) {
		// zh on its own is China, but traditional chinese is Taiwan
		code, err := currency.GetCurrencyForLocale("zh-Hant")
		assert.NoError(t, err)
		assert.Equal(t, "TWD", code)

		code, err = currency.GetCurrencyForLocale("zh-Hans")
		assert.NoError(t, err)
		assert.Equal(t, "CNY", code)
	})

	t.Run("unicode extension is not a region", func(t *testing.T) {
		// nu is also the region code for Niue, this used to return NZD
		code, err := currency.GetCurrencyForLocale("en-u-nu-latn")
		assert.NoError(t, err)
		assert.Equal(t, "USD", code)
	})

	t.Run("region without a currency", func(t *testing.T) {
		// 419 is Latin America which isn't a single country
		code, err := currency.GetCurrencyForLocale("es-419")
		assert.EqualError(t, err, "no currency found for locale [es-419] region [419]")
		assert.Empty(t, code)
	})

	t.Run("macro region without a currency", func(t *testing.T) {
		// 150 is Europe, which isn't a single country either
		code, err := currency.GetCurrencyForLocale("en-150")
		assert.EqualError(t, err, "no currency found for locale [en-150] region [150]")
		assert.Empty(t, code)
	})

	t.Run("unknown language", func(t *testing.T) {
		code, err := currency.GetCurrencyForLocale("xx")
		assert.Error(t, err)
		assert.Empty(t, code)
	})

	t.Run("blank", func(t *testing.T) {
		code, err := currency.GetCurrencyForLocale("")
		assert.EqualError(t, err, "locale cannot be blank")
		assert.Empty(t, code)
	})
}

func TestMatchLocale(t *testing.T) {
	t.Run("falls back to english", func(t *testing.T) {
		for _, header := range []string{"", "!!", "*", "xx", "en;q=0"} {
			assert.Equal(t, "en", currency.MatchLocale(header), "should fall back to en for %q", header)
		}
	})

	t.Run("picks the closest locale", func(t *testing.T) {
		cases := map[string]string{
			"en-US,en;q=0.9": "en",
			"de-DE,de;q=0.9": "de",
			"de-CH":          "de-CH",
			"fr-CA":          "fr-CA",
			"ja":             "ja",
			"zh-TW":          "zh-Hant",
		}
		for header, expected := range cases {
			assert.Equal(t, expected, currency.MatchLocale(header), "wrong locale for %q", header)
		}
	})
}

func TestGetCurrency(t *testing.T) {
	t.Run("english", func(t *testing.T) {
		result, err := currency.GetCurrency("en", "USD")
		assert.NoError(t, err)
		assert.Equal(t, currency.Currency{
			Code:             "USD",
			Name:             "US Dollar",
			Symbol:           "$",
			DecimalSeparator: ".",
			GroupSeparator:   ",",
			MinusSign:        "-",
			FractionalDigits: 2,
		}, result)
	})

	t.Run("japanese", func(t *testing.T) {
		result, err := currency.GetCurrency("ja", "JPY")
		assert.NoError(t, err)
		assert.Equal(t, currency.Currency{
			Code:             "JPY",
			Name:             "日本円",
			Symbol:           "￥",
			DecimalSeparator: ".",
			GroupSeparator:   ",",
			MinusSign:        "-",
			FractionalDigits: 0,
		}, result)
	})

	t.Run("inherits from the parent locale", func(t *testing.T) {
		// zh-Hant-HK doesn't have its own name for USD so it should come from
		// zh-Hant, not from zh which is simplified chinese
		result, err := currency.GetCurrency("zh-Hant-HK", "USD")
		assert.NoError(t, err)
		assert.Equal(t, "美元", result.Name)
	})

	t.Run("script locale does not inherit other scripts", func(t *testing.T) {
		// az-Cyrl's parent is root and not az, so anything it doesn't have should
		// fall back to english instead of the latin script names from az
		result, err := currency.GetCurrency("az-Cyrl", "AMD")
		assert.NoError(t, err)
		assert.Equal(t, "Armenian Dram", result.Name)

		result, err = currency.GetCurrency("az", "AMD")
		assert.NoError(t, err)
		assert.Equal(t, "Ermənistan Dramı", result.Name)
	})

	t.Run("includes the separators", func(t *testing.T) {
		result, err := currency.GetCurrency("de", "EUR")
		assert.NoError(t, err)
		assert.Equal(t, ",", result.DecimalSeparator)
		assert.Equal(t, ".", result.GroupSeparator)

		// Cape Verde's own currency has its own decimal separator
		result, err = currency.GetCurrency("pt-CV", "CVE")
		assert.NoError(t, err)
		assert.Equal(t, "$", result.DecimalSeparator)
	})

	t.Run("unsupported currency", func(t *testing.T) {
		result, err := currency.GetCurrency("en", "DEM")
		assert.EqualError(t, err, "currency not supported")
		assert.Empty(t, result)
	})
}

func TestGetCurrencyList(t *testing.T) {
	t.Run("has every currency", func(t *testing.T) {
		codes := currency.GetCurrencies()
		result := currency.GetCurrencyList("en")
		assert.Len(t, result, len(codes), "should have every supported currency")
		assert.True(t, slices.IsSortedFunc(result, func(a, b currency.Currency) int {
			if a.Code < b.Code {
				return -1
			} else if a.Code > b.Code {
				return 1
			}
			return 0
		}), "should be sorted by code")
		for _, item := range result {
			assert.NotEmpty(t, item.Name, "should always have a name for %s", item.Code)
			assert.NotEmpty(t, item.Symbol, "should always have a symbol for %s", item.Code)
		}
	})
}

func TestGetFractionalDigits(t *testing.T) {
	t.Run("supported currencies", func(t *testing.T) {
		cases := map[string]int64{
			"USD": 2,
			"EUR": 2,
			"JPY": 0,
			"BHD": 3,
		}
		for code, expected := range cases {
			assert.True(t, currency.IsSupportedCurrency(code), "%s should be supported", code)
			digits, err := currency.GetFractionalDigits(code)
			assert.NoError(t, err)
			assert.EqualValues(t, expected, digits, "wrong digits for %s", code)
		}
	})

	t.Run("retired currency", func(t *testing.T) {
		assert.False(t, currency.IsSupportedCurrency("DEM"), "DEM should not be supported")
		digits, err := currency.GetFractionalDigits("DEM")
		assert.EqualError(t, err, "currency not supported")
		assert.Zero(t, digits)
	})
}

func TestGetMinusSign(t *testing.T) {
	t.Run("minus signs", func(t *testing.T) {
		cases := map[string]string{
			"en":    "-",
			"de":    "-",
			"fi":    "\u2212",
			"hr":    "\u2212",
			"ar-EG": "\u200e-",
			"fa":    "\u200e\u2212",
		}
		for locale, expected := range cases {
			assert.Equal(t, expected, currency.GetMinusSign(locale), "wrong minus sign for %s", locale)
		}
	})
}

func TestGetSeparators(t *testing.T) {
	t.Run("locale separators", func(t *testing.T) {
		cases := []struct {
			locale  string
			code    string
			decimal string
			group   string
		}{
			{"en", "USD", ".", ","},
			{"de", "EUR", ",", "."},
			{"de-CH", "CHF", ".", "'"},
			// This one is a narrow no-break space, not a regular space
			{"fr", "EUR", ",", " "},
			// ar-EG defaults to arabic-indic digits but we only use latin ones
			{"ar-EG", "EGP", ".", ","},
		}
		for _, item := range cases {
			decimal, group := currency.GetSeparators(item.locale, item.code)
			assert.Equal(t, item.decimal, decimal, "wrong decimal for %s %s", item.locale, item.code)
			assert.Equal(t, item.group, group, "wrong group for %s %s", item.locale, item.code)
		}
	})

	t.Run("currency specific separators", func(t *testing.T) {
		// Cape Verde writes escudos like 1$50
		decimal, _ := currency.GetSeparators("pt-CV", "CVE")
		assert.Equal(t, "$", decimal)

		// But other currencies in the same locale still use the normal ones
		decimal, _ = currency.GetSeparators("pt-CV", "EUR")
		assert.Equal(t, ",", decimal)
	})
}

func TestGetCLDRVersion(t *testing.T) {
	t.Run("has a version", func(t *testing.T) {
		assert.NotEmpty(t, currency.GetCLDRVersion())
	})
}
