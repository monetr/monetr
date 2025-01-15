package calc

import (
	"math/big"

	"github.com/pkg/errors"
)

// Deprecated: use the currency package instead
func ConvertStringToCents(input string) (int64, error) {
	f, _, err := big.ParseFloat(input, 10, 64, big.ToNearestEven)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert string amount to cents")
	}
	// Convert to cents
	f = f.Mul(f, big.NewFloat(100))
	amount, _ := f.Int64()

	return amount, nil
}
