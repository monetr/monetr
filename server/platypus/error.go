package platypus

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/v41/plaid"
)

const (
	// ErrorCodeTransactionsSyncMutationDuringPagination is returned by Plaid's
	// /transactions/sync endpoint when the underlying transaction data mutated
	// while we were paginating through it with a stored cursor. Per Plaid's
	// documentation the remediation is to restart pagination from a null cursor.
	ErrorCodeTransactionsSyncMutationDuringPagination = "TRANSACTIONS_SYNC_MUTATION_DURING_PAGINATION"
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

// IsPlaidErrorCode reports whether err, once unwrapped, is a *PlatypusError
// whose ErrorCode matches the provided code.
func IsPlaidErrorCode(err error, code string) bool {
	if err == nil {
		return false
	}
	plaidErr, ok := errors.Cause(err).(*PlatypusError)
	if !ok {
		return false
	}
	return plaidErr.ErrorCode == code
}
