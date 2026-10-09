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
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)
	span.SetData("transactionRecurringId", transactionRecurringId)

	var result TransactionRecurring
	err := r.txn.NewSelect().
		Model(&result).
		Relation("FundingSchedule").
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

func (r *repositoryBase) GetTransactionRecurrings(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	direction *Direction,
	ended *bool,
	limit, offset int,
) ([]TransactionRecurring, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)
	span.SetData("direction", direction)
	span.SetData("ended", ended)

	result := make([]TransactionRecurring, 0)
	query := r.txn.NewSelect().
		Model(&result).
		Relation("FundingSchedule").
		Where(`"transaction_recurring"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_recurring"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_recurring"."deleted_at" IS NULL`)

	if direction != nil {
		query = query.Where(`"transaction_recurring"."direction" = ?`, *direction)
	}

	if ended != nil {
		query = query.Where(`"transaction_recurring"."ended" = ?`, *ended)

		// Ended ones still project next forward off of their rule so it doesn't
		// mean anything for them, go by when they were last seen instead
		if *ended {
			query = query.Order(`transaction_recurring.last DESC`)
		}
	}

	err := query.
		Limit(limit).
		Offset(offset).
		// Active ones first, then whatever is coming up next
		Order(`transaction_recurring.ended ASC`).
		Order(`transaction_recurring.next ASC`).
		Order(`transaction_recurring.transaction_recurring_id DESC`).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve recurring transactions")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetTransactionRecurringByCluster(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionClusterId ID[TransactionCluster],
) ([]TransactionRecurring, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)
	span.SetData("transactionClusterId", transactionClusterId)

	result := make([]TransactionRecurring, 0)
	if err := r.txn.NewSelect().
		Model(&result).
		Where(`"transaction_recurring"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_recurring"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_recurring"."transaction_cluster_id" = ?`, transactionClusterId).
		Scan(span.Context()); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to retrieve recurring transactions for cluster",
		)
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetTransactionRecurringByBankAccount(
	ctx context.Context,
	bankAccountId ID[BankAccount],
) ([]TransactionRecurring, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)

	result := make([]TransactionRecurring, 0)
	if err := r.txn.NewSelect().
		Model(&result).
		Where(`"transaction_recurring"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_recurring"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_recurring"."deleted_at" IS NULL`).
		Scan(span.Context()); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(
			span.Context(),
			err,
			"failed to retrieve recurring transactions for bank account",
		)
	}

	span.Status = sentry.SpanStatusOK

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
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)

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
		span.Status = sentry.SpanStatusInternalError
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to upsert recurring transactions",
		)
	}

	span.Status = sentry.SpanStatusOK

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
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)

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
		span.Status = sentry.SpanStatusInternalError
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to update transaction recurring ids",
		)
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) UpdateTransactionRecurring(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	recurring *TransactionRecurring,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)
	span.SetData("transactionRecurringId", recurring.TransactionRecurringId)

	recurring.AccountId = r.AccountId()
	recurring.UpdatedAt = r.clock.Now()

	_, err := r.txn.NewUpdate().
		Model(recurring).
		Where(`"transaction_recurring"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_recurring"."bank_account_id" = ?`, bankAccountId).
		WherePK().
		Returning("*").
		Exec(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update recurring transaction")
	}

	span.Status = sentry.SpanStatusOK

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
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)

	_, err := r.txn.NewDelete().
		Model(new(TransactionRecurring)).
		Where(`"transaction_recurring"."account_id" = ?`, r.AccountId()).
		Where(`"transaction_recurring"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_recurring"."transaction_recurring_id" IN (?)`, bun.List(transactionRecurringIds)).
		Exec(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return crumbs.WrapError(
			span.Context(),
			err,
			"failed to delete obsolete recurring transactions",
		)
	}

	span.Status = sentry.SpanStatusOK

	return nil
}
