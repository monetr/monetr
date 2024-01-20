package repository

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreatePlaidTransaction(
	ctx context.Context,
	transaction *models.PlaidTransaction,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transaction.AccountId = r.AccountId()
	transaction.CreatedAt = r.clock.Now().UTC()

	r.txn.ModelContext(span.Context(), transaction).Insert(transaction)

	return nil
}

func (r *repositoryBase) GetTransactionByPendingTransactionPlaidId(
	ctx context.Context,
	bankAccountId uint64,
	plaidPendingTransactionId string,
) (*models.Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var txn models.Transaction
	err := r.txn.ModelContext(span.Context(), &txn).
		Join(`INNER JOIN "plaid_transactions" AS "plaid_transaction"`).
		JoinOn(`"plaid_transaction"."plaid_transaction_id" = "transaction"."pending_plaid_transaction_id" AND "plaid_transaction"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"plaid_transaction"."plaid_id" = ?`, plaidPendingTransactionId).
		Limit(1).
		Select(&txn)
	if err == pg.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to find transaction by plaid pending transaction id")
	}

	return &txn, nil
}
