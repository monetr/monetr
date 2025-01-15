package currency

import (
	"math/big"

	locale "github.com/elliotcourant/go-lclocale"
	"github.com/pkg/errors"
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
		return 0, errors.Wrap(err, "failed to parse currency amount")
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
	// Convert that back into a regular int64.
	amount, accuracy := f.Int64()
	if accuracy != big.Exact {
		return amount, errors.Errorf("failed to parse currency amount accurately: %s", accuracy.String())
	}

	return amount, nil
}
