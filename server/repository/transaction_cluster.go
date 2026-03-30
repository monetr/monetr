package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/types"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetTransactionClusterMembersByBankAccount(
	ctx context.Context,
	bankAccountId ID[BankAccount],
) ([]TransactionClusterMember, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result []TransactionClusterMember
	if err := r.txn.ModelContext(
		span.Context(),
		&result,
	).Where(`"transaction_cluster_member"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster_member"."bank_account_id" = ?`, bankAccountId).
		Select(&result); err != nil {
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to retrieve all transaction cluster members",
		)
	}

	return result, nil
}

func (r *repositoryBase) UpsertTransactionClusters(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	clusters []TransactionCluster,
) error {
	if len(clusters) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for i := range clusters {
		clusters[i].AccountId = r.AccountId()
		clusters[i].BankAccountId = bankAccountId
	}

	_, err := r.txn.ModelContext(span.Context(), &clusters).
		OnConflict(`("transaction_cluster_id", "account_id", "bank_account_id") DO UPDATE`).
		Set(`"name" = EXCLUDED."name"`).
		Set(`"original_name" = EXCLUDED."original_name"`).
		Set(`"signature" = EXCLUDED."signature"`).
		Set(`"centroid" = EXCLUDED."centroid"`).
		Set(`"members" = EXCLUDED."members"`).
		Set(`"debug" = EXCLUDED."debug"`).
		Set(`"merchant" = EXCLUDED."merchant"`).
		Set(`"updated_at" = NOW()`).
		Insert()
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to upsert transaction clusters",
		)
	}

	return nil
}

func (r *repositoryBase) DeleteTransactionClusters(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	clusterIds []ID[TransactionCluster],
) error {
	if len(clusterIds) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &TransactionCluster{}).
		Where(`"transaction_cluster"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_cluster"."transaction_cluster_id" IN (?)`, pg.In(clusterIds)).
		Delete()
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to delete obsolete transaction clusters",
		)
	}

	return nil
}

func (r *repositoryBase) UpsertTransactionClusterMembers(
	ctx context.Context,
	members []TransactionClusterMember,
) error {
	if len(members) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &members).
		OnConflict(`("transaction_id", "account_id", "bank_account_id") DO UPDATE`).
		Set(`"transaction_cluster_id" = EXCLUDED."transaction_cluster_id"`).
		Set(`"updated_at" = NOW()`).
		Insert()
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to upsert transaction cluster members",
		)
	}

	return nil
}

func (r *repositoryBase) DeleteTransactionClusterMembers(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionIds []ID[Transaction],
) error {
	if len(transactionIds) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &TransactionClusterMember{}).
		Where(`"transaction_cluster_member"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster_member"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_cluster_member"."transaction_id" IN (?)`, pg.In(transactionIds)).
		Delete()
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to delete transaction cluster members",
		)
	}

	return nil
}

func (r *repositoryBase) WriteTransactionClusters(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	clusters []TransactionCluster,
) (updated []TransactionCluster, err error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Build an array of signature + centroid pairs that we want to keep. We will
	// delete verything that isn't in this dataset.
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

	var deleted []TransactionCluster
	// Because transaction clusters are calculated using the entire bank accounts
	// transaction dataset. If any clusters exist right now that don't have the
	// same signature/centroid pair then we can safely remove them. Either those
	// transactions were part of a lower quality cluster, or the transactions were
	// merged into a newer larger cluster based on updated information. Either way
	// the old cluster needs to be removed. This might be volatile for accounts
	// with less data, but becomes very stable the more data the accounts have.
	cleanResult, err := r.txn.ModelContext(span.Context(), &TransactionCluster{}).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		WhereIn(`("signature", "centroid") NOT IN (?)`, keysToKeep).
		Returning("*").
		Delete(&deleted)
	if err != nil {
		return nil, errors.Wrap(err, "failed to clean up outdated transaction clusters")
	}

	// Then we can insert all the transaction clusters we have calculated. But if
	// we get a conflict on our unique index then we will instead just update the
	// existing cluster.
	result, err := r.txn.ModelContext(span.Context(), &clusters).
		OnConflict(`("account_id", "bank_account_id", "signature", "centroid") DO UPDATE`).
		Set(`"members" = EXCLUDED.members`).
		Set(`"debug" = EXCLUDED.debug`).
		Set(`"merchant" = EXCLUDED.merchant`).
		// TODO It is possible for a cluster to be recalculated with no changes
		// whatsoever. When this happens it does not exactly make sense to update
		// the updated_at timestamp here. But it would involve pulling that clusters
		// data from the database and comparing it in someway.
		Set(`"updated_at" = now()`).
		Where(`"transaction_cluster"."members" != EXCLUDED.members`).
		// Return rows that we touched. This will only return rows that were either
		// freshly inserted or were updated by this upsert. Rows that are identical
		// to existing data are not returned here.
		Returning(`"transaction_cluster".*`).
		Insert(&clusters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert the new transaction clusters")
	}

	if len(deleted) > 0 {
		r.log.Log(
			span.Context(),
			logging.LevelTrace,
			"cleaned outdated transaction clusters",
			"removed", deleted,
			"upserted", clusters,
		)
	}

	r.log.DebugContext(
		span.Context(),
		"upserted transaction clusters",
		"cleaned", cleanResult.RowsAffected(),
		"returned", result.RowsReturned(),
		"affected", result.RowsAffected(),
	)

	return clusters, nil
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

func (r *repositoryBase) GetTransactionsByCluster(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionClusterId ID[TransactionCluster],
	limit, offset int,
) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":            r.AccountId(),
		"bankAccountId":        bankAccountId,
		"transactionClusterId": transactionClusterId,
		"limit":                limit,
		"offset":               offset,
	}

	items := make([]Transaction, 0)
	err := r.txn.ModelContext(span.Context(), &items).
		Join(`INNER JOIN "transaction_cluster_members" AS "transaction_cluster_member"`).
		JoinOn(`"transaction_cluster_member"."transaction_id" = "transaction"."transaction_id"`).
		JoinOn(`"transaction_cluster_member"."bank_account_id" = "transaction"."bank_account_id"`).
		JoinOn(`"transaction_cluster_member"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."deleted_at" IS NULL`).
		Where(`"transaction_cluster_member"."transaction_cluster_id" = ?`, transactionClusterId).
		Limit(limit).
		Offset(offset).
		Order(`date DESC`).
		Order(`transaction_id DESC`).
		Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(span.Context(), err, "failed to retrieve transactions")
	}

	span.Status = sentry.SpanStatusOK

	return items, nil

}
