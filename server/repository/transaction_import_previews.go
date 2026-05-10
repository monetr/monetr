package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetTransactionImportPreviewByTransactionImportId(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionImportId ID[TransactionImport],
) (*TransactionImportPreview, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var item TransactionImportPreview
	err := r.txn.ModelContext(span.Context(), &item).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_import_id" = ?`, transactionImportId).
		Limit(1).
		Select(&item)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction import preview")
	}

	span.Status = sentry.SpanStatusOK

	return &item, nil
}

func (r *repositoryBase) CreateTransactionImportPreview(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionImportPreview *TransactionImportPreview,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transactionImportPreview.AccountId = r.AccountId()
	transactionImportPreview.BankAccountId = bankAccountId
	transactionImportPreview.CreatedAt = r.clock.Now().UTC()
	_, err := r.txn.ModelContext(span.Context(), transactionImportPreview).
		Insert(transactionImportPreview)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction import record")
	}

	span.SetData(
		"transactionImportPreviewId",
		transactionImportPreview.TransactionImportId.String(),
	)
	span.Status = sentry.SpanStatusOK

	return nil
}
