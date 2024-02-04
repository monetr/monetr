package teller

import (
	"strconv"

	"github.com/pkg/errors"
)

type Balance struct {
	AccountId string            `json:"account_id"`
	Ledger    string            `json:"ledger"`
	Available string            `json:"available"`
	Links     map[string]string `json:"links"`
}

func (b Balance) GetLedger() (int64, error) {
	ledger, err := strconv.ParseFloat(b.Ledger, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse ledger balance")
	}

	// Convert to total cents
	return int64(ledger) * 100, nil
}

// GetAvailable will return the available balance from Teller's API as a 64 bit
// integer representing the total cents of the balance instead of a floating
// point number.
func (b Balance) GetAvailable() (int64, error) {
	balance, err := strconv.ParseFloat(b.Available, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse available balance")
	}

	// Convert to total cents
	return int64(balance) * 100, nil
}
