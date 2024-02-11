package teller

import (
	"github.com/monetr/monetr/server/internal/calc"
	"github.com/pkg/errors"
)

type Balance struct {
	AccountId string            `json:"account_id"`
	Ledger    string            `json:"ledger"`
	Available string            `json:"available"`
	Links     map[string]string `json:"links"`
}

func (b Balance) GetLedger() (int64, error) {
	balance, err := calc.ConvertStringToCents(b.Ledger)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse ledger balance")
	}

	return balance, nil
}

// GetAvailable will return the available balance from Teller's API as a 64 bit
// integer representing the total cents of the balance instead of a floating
// point number.
func (b Balance) GetAvailable() (int64, error) {
	balance, err := calc.ConvertStringToCents(b.Available)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse available balance")
	}

	return balance, nil
}
