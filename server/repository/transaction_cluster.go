package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/uptrace/bun"
)

func (r *repositoryBase) GetTransactionClusterMembersByBankAccount(
	ctx context.Context,
	bankAccountId ID[BankAccount],
) ([]TransactionClusterMember, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result []TransactionClusterMember
	if err := r.txn.NewSelect().
		Model(&result).
		Where(`"transaction_cluster_member"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster_member"."bank_account_id" = ?`, bankAccountId).
		Scan(span.Context()); err != nil {
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

	now := r.clock.Now()
	for i := range clusters {
		clusters[i].AccountId = r.AccountId()
		clusters[i].BankAccountId = bankAccountId
		clusters[i].UpdatedAt = now
	}

	_, err := r.txn.NewInsert().
		Model(&clusters).
		On(`CONFLICT ("transaction_cluster_id", "account_id", "bank_account_id") DO UPDATE`).
		Set(`"original_name" = EXCLUDED."original_name"`).
		Set(`"signature" = EXCLUDED."signature"`).
		Set(`"centroid" = EXCLUDED."centroid"`).
		Set(`"members" = EXCLUDED."members"`).
		Set(`"debug" = EXCLUDED."debug"`).
		Set(`"merchant" = EXCLUDED."merchant"`).
		Set(`"updated_at" = EXCLUDED."updated_at"`).
		Exec(span.Context())
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

	_, err := r.txn.NewDelete().
		Model(&TransactionCluster{}).
		Where(`"transaction_cluster"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_cluster"."transaction_cluster_id" IN (?)`, bun.In(clusterIds)).
		Exec(span.Context())
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

	now := r.clock.Now()
	for i := range members {
		members[i].AccountId = r.AccountId()
		members[i].UpdatedAt = now
	}

	_, err := r.txn.NewInsert().
		Model(&members).
		On(`CONFLICT ("transaction_id", "account_id", "bank_account_id") DO UPDATE`).
		Set(`"transaction_cluster_id" = EXCLUDED."transaction_cluster_id"`).
		Set(`"updated_at" = EXCLUDED."updated_at"`).
		Exec(span.Context())
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

	_, err := r.txn.NewDelete().
		Model(&TransactionClusterMember{}).
		Where(`"transaction_cluster_member"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster_member"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_cluster_member"."transaction_id" IN (?)`, bun.In(transactionIds)).
		Exec(span.Context())
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to delete transaction cluster members",
		)
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
	err := r.txn.NewSelect().
		Model(&cluster).
		Where(`"transaction_cluster"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster"."bank_account_id" = ?`, bankAccountId).
		Where(`? = ANY ("transaction_cluster"."members")`, transactionId).
		Limit(1).
		Scan(span.Context())
	if err != nil {
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to find cluster containing transaction",
		)
	}

	return &cluster, nil
}

func (r *repositoryBase) GetTransactionCluster(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionClusterId ID[TransactionCluster],
) (*TransactionCluster, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":            r.AccountId(),
		"bankAccountId":        bankAccountId,
		"transactionClusterId": transactionClusterId,
	}

	var result TransactionCluster
	err := r.txn.NewSelect().
		Model(&result).
		Where(`"transaction_cluster"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_cluster"."transaction_cluster_id" = ?`, transactionClusterId).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to retrieve transaction cluster",
		)
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
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
	err := r.txn.NewSelect().
		Model(&items).
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
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(span.Context(), err, "failed to retrieve transactions")
	}

	span.Status = sentry.SpanStatusOK

	return items, nil

}
