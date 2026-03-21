package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreatePlaidTransactions(
	ctx context.Context,
	transactions ...*PlaidTransaction,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
		transactions[i].CreatedAt = r.clock.Now().UTC()
	}

	_, err := r.txn.ModelContext(span.Context(), &transactions).Insert(&transactions)
	return errors.Wrap(err, "failed to insert plaid transactions")
}
