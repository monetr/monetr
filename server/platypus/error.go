package platypus

import (
	"fmt"

	"github.com/plaid/plaid-go/v30/plaid"
)

var (
	_ error = &PlatypusError{}
)

type PlatypusError struct {
	plaid.PlaidError
}

func (p *PlatypusError) Error() string {
	return fmt.Sprintf(
		"plaid API call failed with [%s - %s]%s",
		p.ErrorType, p.ErrorCode, p.ErrorMessage,
	)
}
