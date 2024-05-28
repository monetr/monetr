package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetTransactionUpload(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionUploadId ID[TransactionUpload],
) (*TransactionUpload, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var item TransactionUpload
	err := r.txn.ModelContext(span.Context(), &item).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Where(`"transaction_upload_id" = ?`, transactionUploadId).
		Limit(1).
		Select(&item)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction upload")
	}

	span.Status = sentry.SpanStatusOK

	return &item, nil
}

func (r *repositoryBase) CreateTransactionUpload(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionUpload *TransactionUpload,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transactionUpload.AccountId = r.AccountId()
	transactionUpload.BankAccountId = bankAccountId
	transactionUpload.CreatedAt = r.clock.Now().UTC()
	transactionUpload.CreatedBy = r.UserId()

	_, err := r.txn.ModelContext(span.Context(), transactionUpload).
		Insert(transactionUpload)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction upload record")
	}

	span.Data["transactionUploadId"] = transactionUpload.TransactionUploadId
	span.Status = sentry.SpanStatusOK

	return nil
}
