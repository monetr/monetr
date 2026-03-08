package repository

import (
	"context"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/types"
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

	keysToKeep := []types.ValueAppender{}

	for i := range clusters {
		cluster := clusters[i]
		cluster.AccountId = r.AccountId()
		cluster.BankAccountId = bankAccountId

		if cluster.Centroid != nil {
			keysToKeep = append(keysToKeep, pg.InMulti([]string{
				cluster.Signature,
				string(*cluster.Centroid),
			}))
		}
		clusters[i] = cluster
	}

	// Because transaction clusters are calculated using the entire bank accounts
	// transaction dataset. If any clusters exist right now that don't have the
	// same signature/centroid pair then we can safely remove them. Either those
	// transactions were part of a lower quality cluster, or the transactions were
	// merged into a newer larger cluster based on updated information. Either way
	// the old cluster needs to be removed. This might be volatile for accounts
	// with less data, but becomes very stable the more data the accounts have.
	_, err := r.txn.ModelContext(span.Context(), &TransactionCluster{}).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		WhereIn(`("signature", "centroid") NOT IN (?)`, keysToKeep).
		Delete()
	if err != nil {
		return errors.Wrap(err, "failed to clean up outdated transaction clusters")
	}

	// Then we can insert all the transaction clusters we have calculated. But if
	// we get a conflict on our unique index then we will instead just update the
	// existing cluster.
	_, err = r.txn.ModelContext(span.Context(), &clusters).
		OnConflict(`("account_id", "bank_account_id", "signature", "centroid") DO UPDATE`).
		Set(`members = EXCLUDED.members`).
		Set(`debug = EXCLUDED.debug`).
		// TODO It is possible for a cluster to be recalculated with no changes
		// whatsoever. When this happens it does not exactly make sense to update
		// the updated_at timestamp here. But it would involve pulling that clusters
		// data from the database and comparing it in someway.
		Set(`updated_at = now()`).
		Insert(&clusters)
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
