package platypus

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/v41/plaid"
	"github.com/stretchr/testify/assert"
)

func TestIsPlaidErrorCode(t *testing.T) {
	t.Run("matches wrapped PlatypusError", func(t *testing.T) {
		err := errors.Wrap(
			&PlatypusError{PlaidError: plaid.PlaidError{
				ErrorType:    "TRANSACTIONS_ERROR",
				ErrorCode:    ErrorCodeTransactionsSyncMutationDuringPagination,
				ErrorMessage: "Underlying transaction data changed since last page was fetched. Please restart pagination from last update.",
			}},
			"failed to sync data with Plaid",
		)
		assert.True(
			t,
			IsPlaidErrorCode(err, ErrorCodeTransactionsSyncMutationDuringPagination),
			"must extract plaid error code through errors.Wrap",
		)
	})

	t.Run("does not match non-plaid error", func(t *testing.T) {
		err := errors.Wrap(errors.New("boom"), "something went wrong")
		assert.False(
			t,
			IsPlaidErrorCode(err, ErrorCodeTransactionsSyncMutationDuringPagination),
			"non-plaid errors must not match",
		)
	})

	t.Run("does not match different code", func(t *testing.T) {
		err := errors.Wrap(
			&PlatypusError{PlaidError: plaid.PlaidError{
				ErrorType:    "ITEM_ERROR",
				ErrorCode:    "ITEM_LOGIN_REQUIRED",
				ErrorMessage: "the user needs to re-auth",
			}},
			"failed to sync data with Plaid",
		)
		assert.False(
			t,
			IsPlaidErrorCode(err, ErrorCodeTransactionsSyncMutationDuringPagination),
			"different plaid error codes must not match",
		)
	})

	t.Run("nil error returns false", func(t *testing.T) {
		assert.False(
			t,
			IsPlaidErrorCode(nil, ErrorCodeTransactionsSyncMutationDuringPagination),
			"nil error must not match",
		)
	})
}
