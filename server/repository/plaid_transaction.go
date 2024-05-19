package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreatePlaidTransaction(
	ctx context.Context,
	transaction *PlaidTransaction,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transaction.AccountId = r.AccountId()
	transaction.CreatedAt = r.clock.Now().UTC()

	r.txn.ModelContext(span.Context(), transaction).Insert(transaction)

	return nil
}

func (r *repositoryBase) DeletePlaidTransaction(
	ctx context.Context,
	plaidTransactionId ID[PlaidTransaction],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &PlaidTransaction{}).
		Set(`"deleted_at" = ?`, r.clock.Now().UTC()).
		Where(`"plaid_transaction"."account_id" = ?`, r.AccountId()).
		Where(`"plaid_transaction"."plaid_transaction_id" = ?`, plaidTransactionId).
		Update()

	return errors.Wrap(err, "failed to delete plaid transaction")
}
