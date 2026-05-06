package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetTransactionImport(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionImportId ID[TransactionImport],
) (*TransactionImport, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var item TransactionImport
	err := r.txn.ModelContext(span.Context(), &item).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_import_id" = ?`, transactionImportId).
		Limit(1).
		Select(&item)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction import")
	}

	span.Status = sentry.SpanStatusOK

	return &item, nil
}

func (r *repositoryBase) CreateTransactionImport(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionImport *TransactionImport,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transactionImport.AccountId = r.AccountId()
	transactionImport.BankAccountId = bankAccountId
	transactionImport.CreatedAt = r.clock.Now().UTC()
	transactionImport.CreatedBy = r.UserId()

	_, err := r.txn.ModelContext(span.Context(), transactionImport).
		Insert(transactionImport)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction import record")
	}

	span.SetData("transactionImportId", transactionImport.TransactionImportId.String())
	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) UpdateTransactionImport(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionImport *TransactionImport,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transactionImport.AccountId = r.AccountId()
	transactionImport.BankAccountId = bankAccountId
	transactionImport.UpdatedAt = r.clock.Now().UTC()

	_, err := r.txn.ModelContext(span.Context(), transactionImport).
		WherePK().
		Update(transactionImport)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update transaction import")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}
