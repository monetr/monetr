package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
)

func (r *repositoryBase) CreateTellerTransaction(ctx context.Context, transaction *models.TellerTransaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transaction.AccountId = r.AccountId()
	transaction.CreatedAt = r.clock.Now().UTC()
	transaction.UpdatedAt = r.clock.Now().UTC()

	r.txn.ModelContext(span.Context(), transaction).Insert(transaction)

	return nil
}
