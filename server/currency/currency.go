package currency

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/pkg/errors"
)

var (
	biggestInt = big.NewFloat(math.MaxInt64)
)

// ParseCurrency parses a numeric string into the int64 value that monetr
// ultimately stores for currency. This is done without using bigint or floating
// point numbers as they can cause odd behaviors. This parser is based on
// PostgreSQL's cash_in implementation here:
// https://github.com/postgres/postgres/blob/801b4ee7fae1caa962b789e72be11dcead79dcbf/src/backend/utils/adt/cash.c#L173
// However this implementation does not handle currency symbols
func ParseCurrency(input string, currency string) (int64, error) {
	// TODO This should be outside the parsing implementation, the parser should
	// instead just receive this information from the caller.
	// Retrieve the fractional digits for the currency we are parsing.
	fractionalDigits, err := locale.GetCurrencyInternationalFractionalDigits(currency)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get currency information [%s]", currency)
	}

	var value int64
	var decimal int64

	// These symbols would normally be derived from locale data. PostgreSQL uses
	// the data from `locales.h` to get this data. But I want to instead derive
	// the data from CLDR long term, for example:
	// https://github.com/unicode-org/cldr-json/blob/c20dc7d37c1080addee64d4f94fbeabbc34f0cc3/cldr-json/cldr-numbers-full/main/nl/numbers.json#L14
	decimalSymbol := "."
	seperatorSymbol := ","
	currencySymbol := "" // Blank for now, eventually should be from locale.
	positiveSymbol := "" // Blank for now. Defaults to positive.
	negativeSymbol := "-"

	sign := 1
	seenDot := false
	// Copy the input string because we are going to modify it
	str := strings.TrimSpace(input)

ParseLoop:
	for ; len(str) > 0; str = str[1:] {
		// Get the current character for the iterator
		character := rune(str[0])
		switch {
		case unicode.IsDigit(character) && (!seenDot || decimal < fractionalDigits):
			// We can calculate our digit by subtracting it from the byte '0'.
			digit := int64(character - '0')
			// TODO: Implement overflow checks.
			// In PostgreSQL this multiplication is done with overflow checking, but
			// I've omitted it here because it introduced complex logic. While it is
			// more correct to have it I want to find an elegant way to include it.
			value *= 10
			value -= digit

			// If we have seen the decimal symbol then we should increment the
			// decimal place counter.
			if seenDot {
				decimal++
			}
		case strings.HasPrefix(str, decimalSymbol) && !seenDot:
			// If the next part of the string has our decimial symbol and we have not
			// already seen a decimal symbol then we need to note this.
			seenDot = true
		case strings.HasPrefix(str, seperatorSymbol):
			// If the next part of the string is the thousands separator then we
			// should consume it and continue to parse.
			str = strings.TrimPrefix(str, seperatorSymbol)
		default:
			break ParseLoop
		}
	}

	// Round up on a thousandth decimal place.
	if len(str) > 0 && unicode.IsDigit(rune(str[0])) && rune(str[0]) > '5' {
		value -= 1
	}

	// Adjust for less than required decimal places
	for ; decimal < fractionalDigits; decimal++ {
		value *= 10
	}

	// Consume any trailing digits in the string
	for ; len(str) > 0 && unicode.IsDigit(rune(str[0])); str = str[1:] {
		// All this loop does is consume trailing digits.
	}

	for len(str) > 0 {
		switch {
		case unicode.IsSpace(rune(str[0])) || rune(str[0]) == ')':
			str = str[1:]
		case strings.HasPrefix(str, negativeSymbol):
			sign = -1
			str = strings.TrimPrefix(str, negativeSymbol)
		case len(positiveSymbol) > 0 && strings.HasPrefix(str, positiveSymbol):
			str = strings.TrimPrefix(str, positiveSymbol)
		case strings.HasPrefix(str, currencySymbol):
			str = strings.TrimPrefix(str, currencySymbol)
		default:
			return 0, errors.Errorf("failed to parse currency %s - %s, unexpected character %s", input, currency, string(str[0]))
		}
	}

	if sign > 0 {
		return value * -1, nil
	} else {
		return value, nil
	}
}

// ParseFriendlyToAmount takes a floating point or whole number as a string,
// this number should not have any thousands separators or currency symbols. But
// it may have the negative symbol preceding the number. Support for more
// flexible inputs will come at a later time.
// This function will then take that amount plus a currency code and return an
// int64 representing that amount in the smallest unit of that currency.
func ParseFriendlyToAmount(
	input string,
	currency string,
) (int64, error) {
	f, _, err := big.ParseFloat(input, 10, 64, big.ToNearestEven)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert string amount to big float")
	}

	// Retrieve the fractional digits for the currency we are parsing.
	fractionalDigits, err := locale.GetCurrencyInternationalFractionalDigits(currency)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get currency information [%s]", currency)
	}

	// Determine the modifier by how many deciml pplaces the currency has. For
	// example, USD has 2 fractional digits, so this would be 10^2 or 100. Where
	// as a currency like JPY would have 0 fractional digits, meaning this ends up
	// being 0.
	base := big.NewInt(10)
	modifier := (new(big.Float)).SetInt(
		// Set the modifier to the result of the exponent.
		base.Exp(base, big.NewInt(fractionalDigits), nil),
	)

	// Convert the provided amount to a whole number representing the smallest
	// unit for that currency.
	f = f.Mul(f, modifier)
	str := f.String()
	parts := strings.Split(str, ".")
	switch {
	case len(parts) == 2 && fractionalDigits == 0:
		return 0, errors.Errorf("invalid input for currency provided, cannot have more than [%d] fractional digits, input: [%s], result: [%s]", fractionalDigits, input, str)
	case f.Cmp(biggestInt) == 1:
		return 0, errors.Errorf("overflow, result is larger than a 64-bit integer: [%s]", str)
	}

	// Convert that back into a regular int64.
	// This is a really stupid approach, but we have basically gaurenteed there
	// would not be a rounding error using the math above. But when we go from a
	// float back to ANY INTEGER EVEN ANOTHER BIG INT it can fuck it up. floating
	// point numbers are the dumbest thing in the entire world.
	amount, _ := strconv.ParseInt(str, 10, 64)
	return amount, nil
}

// ParseFloatToAmount takes a float32 or float64 and converts it into the amount
// that monetr uses and stores for the currency the amount is in. This is done
// in a hacky way where we just convert the float to a string and then parse it
// manually. But it prevents weird rounding errors.
func ParseFloatToAmount[T float32 | float64](
	input T,
	currency string,
) (int64, error) {
	return ParseFriendlyToAmount(fmt.Sprint(input), currency)
}
