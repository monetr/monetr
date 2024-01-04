package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type fileRepositoryInterface interface {
	CreateFile(ctx context.Context, file *models.File) error
	GetFiles(ctx context.Context, bankAccountId uint64) ([]models.File, error)
}

func (r *repositoryBase) GetFiles(ctx context.Context, bankAccountId uint64) ([]models.File, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
	}

	var items []models.File
	err := r.txn.ModelContext(span.Context(), &items).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		Limit(50).
		Order(`file_id DESC`).
		Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve files")
	}

	span.Status = sentry.SpanStatusOK

	return items, nil
}

func (r *repositoryBase) CreateFile(ctx context.Context, file *models.File) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": file.BankAccountId,
	}

	file.AccountId = r.AccountId()
	file.CreatedAt = r.clock.Now().UTC()
	file.CreatedByUserId = r.UserId()

	_, err := r.txn.ModelContext(span.Context(), file).
		Insert(file)
	if err != nil {
		return errors.Wrap(err, "failed to create file record")
	}

	span.Data["spendingId"] = file.FileId
	span.Status = sentry.SpanStatusOK

	return nil

}
