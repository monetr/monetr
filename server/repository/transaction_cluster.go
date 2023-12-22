package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) WriteTransactionClusters(
	ctx context.Context,
	bankAccountId uint64,
	clusters []models.TransactionCluster,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for i := range clusters {
		clusters[i].AccountId = r.AccountId()
		clusters[i].BankAccountId = bankAccountId
	}

	_, err := r.txn.ModelContext(span.Context(), new(models.TransactionCluster)).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Delete()
	if err != nil {
		return errors.Wrap(err, "failed to delete existing transaction clusters")
	}

	_, err = r.txn.ModelContext(span.Context(), &clusters).Insert(&clusters)
	if err != nil {
		return errors.Wrap(err, "failed to insert the new transaction clusters")
	}

	return nil
}
