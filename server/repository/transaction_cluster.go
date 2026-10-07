package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/uptrace/bun"
)

func (r *repositoryBase) GetTransactionClusters(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	limit, offset int,
) ([]TransactionCluster, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"limit":         limit,
		"offset":        offset,
	}

	result := make([]TransactionCluster, 0)
	if err := r.txn.NewSelect().
		Model(&result).
		Where(`"transaction_cluster"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster"."bank_account_id" = ?`, bankAccountId).
		Limit(limit).
		Offset(offset).
		// TODO Figure out some better ordering for this stuff
		Order(`name DESC`).
		Order(`transaction_cluster_id DESC`).
		Scan(span.Context()); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to read transaction clusters",
		)
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetClusteredTransactions(
	ctx context.Context,
	bankAccountId ID[BankAccount],
) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Include soft deleted transactions on purpose, that way the clustering job
	// can pull them out of their cluster
	var result []Transaction
	if err := r.txn.NewSelect().
		Model(&result).
		Column(
			"transaction_id",
			"account_id",
			"bank_account_id",
			"transaction_cluster_id",
		).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."transaction_cluster_id" IS NOT NULL`).
		Scan(span.Context()); err != nil {
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to retrieve clustered transactions",
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
		Set(`"original_memo" = EXCLUDED."original_memo"`).
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
		Where(`"transaction_cluster"."transaction_cluster_id" IN (?)`, bun.List(clusterIds)).
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

func (r *repositoryBase) UpdateTransactionClusterIds(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactions []Transaction,
) error {
	if len(transactions) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
		transactions[i].BankAccountId = bankAccountId
	}

	_, err := r.txn.NewUpdate().
		Model(&transactions).
		Column("transaction_cluster_id").
		Bulk().
		Exec(span.Context())
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to update transaction cluster ids",
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

func (r *repositoryBase) GetTransactionClusterIds(
	ctx context.Context,
	bankAccountId ID[BankAccount],
) ([]ID[TransactionCluster], error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
	}

	result := make([]ID[TransactionCluster], 0)
	err := r.txn.NewSelect().
		Model(new(TransactionCluster)).
		Column("transaction_cluster_id").
		Where(`"transaction_cluster"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_cluster"."bank_account_id" = ?`, bankAccountId).
		Order("transaction_cluster_id").
		Scan(span.Context(), &result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to retrieve transaction cluster ids",
		)
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
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
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."transaction_cluster_id" = ?`, transactionClusterId).
		Where(`"transaction"."deleted_at" IS NULL`).
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
