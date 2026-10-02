package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func (r *repositoryBase) GetTransactionRecurringById(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionRecurringId ID[TransactionRecurring],
) (*TransactionRecurring, error) {

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":              r.AccountId(),
		"bankAccountId":          bankAccountId,
		"transactionRecurringId": transactionRecurringId,
	}

	var result TransactionRecurring
	err := r.txn.NewSelect().
		Model(&result).
		Where(`"transaction_recurring"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_recurring"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_recurring"."transaction_recurring_id" = ?`, transactionRecurringId).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve recurring transaction")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}

func (r *repositoryBase) GetTransactionRecurringByCluster(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionClusterId ID[TransactionCluster],
) ([]TransactionRecurring, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":            r.AccountId(),
		"bankAccountId":        bankAccountId,
		"transactionClusterId": transactionClusterId,
	}

	result := make([]TransactionRecurring, 0)
	if err := r.txn.NewSelect().
		Model(&result).
		Where(`"transaction_recurring"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_recurring"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_recurring"."transaction_cluster_id" = ?`, transactionClusterId).
		Scan(span.Context()); err != nil {
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to retrieve recurring transactions for cluster",
		)
	}

	return result, nil
}

func (r *repositoryBase) UpsertTransactionRecurring(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	recurring []TransactionRecurring,
) error {
	if len(recurring) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	now := r.clock.Now()
	for i := range recurring {
		recurring[i].AccountId = r.AccountId()
		recurring[i].BankAccountId = bankAccountId
		recurring[i].UpdatedAt = now
	}

	// There is only ever one recurring transaction per cluster and direction, so
	// if one already exists then it is updated in place and keeps its ID.
	_, err := r.txn.NewInsert().
		Model(&recurring).
		On(`CONFLICT ("account_id", "bank_account_id", "transaction_cluster_id", "direction") DO UPDATE`).
		Set(`"window_type" = EXCLUDED."window_type"`).
		Set(`"ruleset" = EXCLUDED."ruleset"`).
		Set(`"first" = EXCLUDED."first"`).
		Set(`"last" = EXCLUDED."last"`).
		Set(`"next" = EXCLUDED."next"`).
		Set(`"ended" = EXCLUDED."ended"`).
		Set(`"confidence" = EXCLUDED."confidence"`).
		Set(`"amounts" = EXCLUDED."amounts"`).
		Set(`"last_amount" = EXCLUDED."last_amount"`).
		Set(`"updated_at" = EXCLUDED."updated_at"`).
		Exec(span.Context())
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to upsert recurring transactions",
		)
	}

	return nil
}

func (r *repositoryBase) UpdateTransactionRecurringIds(
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
		Column("transaction_recurring_id").
		Bulk().
		Exec(span.Context())
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to update transaction recurring ids",
		)
	}

	return nil
}

func (r *repositoryBase) DeleteTransactionRecurring(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionRecurringIds []ID[TransactionRecurring],
) error {
	if len(transactionRecurringIds) == 0 {
		return nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.NewDelete().
		Model(&TransactionRecurring{}).
		Where(`"transaction_recurring"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_recurring"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_recurring"."transaction_recurring_id" IN (?)`, bun.List(transactionRecurringIds)).
		Exec(span.Context())
	if err != nil {
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to delete obsolete recurring transactions",
		)
	}

	return nil
}
