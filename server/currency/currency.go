package currency

import (
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"
)

// ParseCurrency parses a numeric string into the int64 value that monetr
// ultimately stores for currency. This is done without using bigint or floating
// point numbers as they can cause odd behaviors. This parser is based on
// PostgreSQL's cash_in implementation here:
// https://github.com/postgres/postgres/blob/801b4ee7fae1caa962b789e72be11dcead79dcbf/src/backend/utils/adt/cash.c#L173
// However this implementation does not handle currency symbols. The decimal
// places and the separators all come from the provided currency, so the same
// input can parse differently depending on the locale the currency came from
func ParseCurrency(input string, currency Currency) (int64, error) {
	fractionalDigits := currency.FractionalDigits

	var value int64
	var decimal int64

	// PostgreSQL gets these from `locales.h` but we get them from CLDR through
	// the currency now. If you are parsing a JSON number then you want a currency
	// from a locale that uses . for the decimal, like en. JSON should only ever
	// use the decimal separator and no thousands separator in numbers. Unless its
	// a string.
	decimalSymbol := currency.DecimalSeparator
	seperatorSymbol := currency.GroupSeparator
	// If the currency doesn't have them then just fall back to what en uses
	if decimalSymbol == "" {
		decimalSymbol = "."
	}
	if seperatorSymbol == "" {
		seperatorSymbol = ","
	}
	currencySymbol := "" // Blank for now, eventually should be from locale.
	// Amounts without a sign are positive, but some banks write an explicit +
	// on their amounts so accept that too.
	positiveSymbol := "+"
	// Always accept a plain - since thats what everyone's keyboard has, but also
	// accept the locale's minus sign (like − in fi). Some locales stick a
	// left-to-right mark in front of it which is really easy to lose when
	// copying, so accept it with or without that too. Longest ones go first
	negativeSymbols := []string{"-"}
	if minus := currency.MinusSign; minus != "" && minus != "-" {
		negativeSymbols = append([]string{
			minus,
			strings.TrimPrefix(minus, "\u200e"),
		}, negativeSymbols...)
	}
	// Returns how long the negative symbol at the start of the string is, or 0
	// if there isn't one
	negativeLength := func(str string) int {
		for _, symbol := range negativeSymbols {
			if symbol != "" && strings.HasPrefix(str, symbol) {
				return len(symbol)
			}
		}
		return 0
	}

	sign := 1
	seenDot := false
	// Blank input, or input that is only a sign like "-", would otherwise just
	// parse as zero. Track whether we actually saw a number so we can error.
	seenDigit := false
	// Copy the input string because we are going to modify it
	str := strings.TrimSpace(input)

	// Consume anything that can come before the digits, same as PostgreSQL. This
	// is where a leading negative sign gets picked up, without this "-5.00" would
	// never get parsed.
PrefixLoop:
	for str != "" {
		switch {
		case str[0] == '(':
			sign = -1
			str = str[1:]
		case negativeLength(str) > 0:
			sign = -1
			str = str[negativeLength(str):]
		case positiveSymbol != "" && strings.HasPrefix(str, positiveSymbol):
			str = strings.TrimPrefix(str, positiveSymbol)
		case currencySymbol != "" && strings.HasPrefix(str, currencySymbol):
			str = strings.TrimPrefix(str, currencySymbol)
		case unicode.IsSpace(rune(str[0])):
			str = str[1:]
		default:
			break PrefixLoop
		}
	}

ParseLoop:
	for ; str != ""; str = str[1:] {
		// Get the current character for the iterator
		character := rune(str[0])
		switch {
		case unicode.IsDigit(character) && (!seenDot || decimal < fractionalDigits):
			seenDigit = true
			// We can calculate our digit by subtracting it from the byte '0'.
			digit := int64(character - '0')
			// PostgreSQL does this multiplication with pg_mul_s64_overflow and
			// pg_sub_s64_overflow, some of the inspiration for this came from there:
			// https://github.com/postgres/postgres/blob/92268b35d04c2de416279f187d12f264afa22614/src/include/common/int.h#L219-L269
			// Those handle any two numbers though, and we only ever multiply by 10
			// and value is always zero or negative. So we can just check value
			// against the smallest number that wont overflow.
			if value < (math.MinInt64+digit)/10 {
				return 0, errors.Errorf(
					"failed to parse currency %s - %s, value is out of range",
					input,
					currency.Code,
				)
			}
			value = value*10 - digit

			// If we have seen the decimal symbol then we should increment the
			// decimal place counter.
			if seenDot {
				decimal++
			}
		case strings.HasPrefix(str, decimalSymbol) && !seenDot:
			// If the next part of the string has our decimial symbol and we have not
			// already seen a decimal symbol then we need to note this.
			seenDot = true
			// The symbols can be more than one byte (like a narrow no-break space) so
			// skip all but the last byte, the loop takes care of the last one.
			str = str[len(decimalSymbol)-1:]
		case strings.HasPrefix(str, seperatorSymbol):
			// If the next part of the string is the thousands separator then we
			// should consume it and continue to parse. Same deal as above, the loop
			// takes care of the last byte.
			str = str[len(seperatorSymbol)-1:]
		default:
			break ParseLoop
		}
	}

	// Round up on a thousandth decimal place.
	if str != "" && unicode.IsDigit(rune(str[0])) && rune(str[0]) >= '5' {
		if value == math.MinInt64 {
			return 0, errors.Errorf(
				"failed to parse currency %s - %s, value is out of range",
				input,
				currency.Code,
			)
		}
		value -= 1
	}

	// Adjust for less than required decimal places
	for ; decimal < fractionalDigits; decimal++ {
		if value < math.MinInt64/10 {
			return 0, errors.Errorf(
				"failed to parse currency %s - %s, value is out of range",
				input,
				currency.Code,
			)
		}
		value *= 10
	}

	// Consume any trailing digits in the string
	for ; str != "" && unicode.IsDigit(rune(str[0])); str = str[1:] {
		// These still count as a number though, like ".5" for JPY which only
		// rounds and never makes it into the parse loop above.
		seenDigit = true
	}

	for str != "" {
		switch {
		case unicode.IsSpace(rune(str[0])) || rune(str[0]) == ')':
			str = str[1:]
		case negativeLength(str) > 0:
			sign = -1
			str = str[negativeLength(str):]
		case positiveSymbol != "" && strings.HasPrefix(str, positiveSymbol):
			str = strings.TrimPrefix(str, positiveSymbol)
		case currencySymbol != "" && strings.HasPrefix(str, currencySymbol):
			// The symbol is blank right now, without checking that this would match
			// everything and spin forever instead of hitting the error below.
			str = strings.TrimPrefix(str, currencySymbol)
		default:
			return 0, errors.Errorf("failed to parse currency %s - %s, unexpected character %s", input, currency.Code, string(str[0]))
		}
	}

	if !seenDigit {
		return 0, errors.Errorf(
			"failed to parse currency %s - %s, no number found",
			input,
			currency.Code,
		)
	}

	if sign > 0 {
		// The negative side of an int64 is one bigger than the positive side, so
		// the smallest number can't be flipped to positive
		if value == math.MinInt64 {
			return 0, errors.Errorf(
				"failed to parse currency %s - %s, value is out of range",
				input,
				currency.Code,
			)
		}
		return value * -1, nil
	} else {
		return value, nil
	}
}

// ParseFloatToAmount takes a float32 or float64 and converts it into the amount
// that monetr uses and stores for the currency the amount is in. This is done
// in a hacky way where we just convert the float to a string and then parse it
// manually. But it prevents weird rounding errors.
func ParseFloatToAmount[T float32 | float64](
	input T,
	currency string,
) (int64, error) {
	fractionalDigits, err := GetFractionalDigits(currency)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get currency information [%s]", currency)
	}

	// Format the float without an exponent, and with the fewest digits that
	// still gets us the same float back. float32 needs its own bit size or
	// something like 0.1 turns into 0.10000000149011612
	var str string
	switch value := any(input).(type) {
	case float32:
		str = strconv.FormatFloat(float64(value), 'f', -1, 32)
	default:
		str = strconv.FormatFloat(float64(input), 'f', -1, 64)
	}

	// A formatted float always uses . for the decimal and never has any grouping,
	// so we don't care what locale the currency would normally be from
	return ParseCurrency(str, Currency{
		Code:             currency,
		DecimalSeparator: ".",
		GroupSeparator:   "",
		MinusSign:        "-",
		FractionalDigits: fractionalDigits,
	})
}

// Currency is a single currency with its name and symbol localized for a
// specific locale
type Currency struct {
	Code             string `json:"code"`
	Name             string `json:"name"`
	Symbol           string `json:"symbol"`
	DecimalSeparator string `json:"decimalSeparator"`
	GroupSeparator   string `json:"groupSeparator"`
	MinusSign        string `json:"minusSign"`
	FractionalDigits int64  `json:"fractionalDigits"`
}
