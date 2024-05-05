package repository

import (
	"context"

	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) WriteTransactionClusters(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	clusters []TransactionCluster,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for i := range clusters {
		clusters[i].AccountId = r.AccountId()
		clusters[i].BankAccountId = bankAccountId
	}

	_, err := r.txn.ModelContext(span.Context(), new(TransactionCluster)).
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

func (r *repositoryBase) GetTransactionClusterByMember(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionId ID[Transaction],
) (*TransactionCluster, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var cluster TransactionCluster
	err := r.txn.ModelContext(span.Context(), &cluster).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Where(`? = ANY ("members")`, transactionId).
		Limit(1).
		Select(&cluster)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find cluster containing transaction")
	}

	return &cluster, nil
}
