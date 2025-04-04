package currency

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/pkg/errors"
)

var (
	biggestInt = big.NewFloat(math.MaxInt64)
)

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
