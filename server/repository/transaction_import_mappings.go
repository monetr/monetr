package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateTransactionImportMapping(
	ctx context.Context,
	mapping *TransactionImportMapping,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	mapping.AccountId = r.AccountId()
	mapping.CreatedBy = r.UserId()
	mapping.CreatedAt = r.clock.Now().UTC()
	mapping.UpdatedAt = r.clock.Now().UTC()

	_, err := r.txn.ModelContext(span.Context(), mapping).
		Insert(mapping)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create transaction import mapping")
	}

	span.SetData("transactionImportMappingId", mapping.TransactionImportMappingId.String())
	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) GetTransactionImportMapping(
	ctx context.Context,
	mappingId ID[TransactionImportMapping],
) (*TransactionImportMapping, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var item TransactionImportMapping
	err := r.txn.ModelContext(span.Context(), &item).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"transaction_import_mapping_id" = ?`, mappingId).
		Limit(1).
		Select(&item)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction import mapping")
	}

	span.Status = sentry.SpanStatusOK

	return &item, nil
}

func (r *repositoryBase) GetTransactionImportMappings(
	ctx context.Context,
	limit, offset int,
) ([]TransactionImportMapping, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]TransactionImportMapping, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"account_id" = ?`, r.AccountId()).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction import mappings")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetTransactionImportMappingsBySignature(
	ctx context.Context,
	signature string,
	limit, offset int,
) ([]TransactionImportMapping, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]TransactionImportMapping, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"signature" = ?`, signature).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction import mappings by signature")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}
