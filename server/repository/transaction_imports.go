package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (r *repositoryBase) GetTransactionImport(
	ctx context.Context,
	linkId ID[Link],
	transactionImportId ID[TransactionImport],
) (*TransactionImport, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var item TransactionImport
	err := r.txn.ModelContext(span.Context(), &item).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"link_id" = ?`, linkId).
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
	linkId ID[Link],
	transactionImport *TransactionImport,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transactionImport.AccountId = r.AccountId()
	transactionImport.LinkId = linkId
	transactionImport.CreatedBy = r.UserId()

	_, err := r.txn.ModelContext(span.Context(), transactionImport).
		Insert(transactionImport)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction import record")
	}

	r.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"linkId":              linkId,
		"transactionImportId": transactionImport.TransactionImportId,
	}).Trace("created transaction import")

	span.SetData("transactionImportId", transactionImport.TransactionImportId.String())
	span.Status = sentry.SpanStatusOK

	return nil
}
