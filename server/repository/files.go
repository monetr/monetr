package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type fileRepositoryInterface interface {
	CreateFile(ctx context.Context, file *File) error
	GetFiles(ctx context.Context) ([]File, error)
}

func (r *repositoryBase) GetFiles(ctx context.Context) ([]File, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
	}

	var items []File
	err := r.txn.ModelContext(span.Context(), &items).
		Where(`"account_id" = ?`, r.AccountId()).
		Limit(100).
		Order(`file_id DESC`).
		Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve files")
	}

	span.Status = sentry.SpanStatusOK

	return items, nil
}

func (r *repositoryBase) CreateFile(ctx context.Context, file *File) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId": r.AccountId(),
	}

	file.AccountId = r.AccountId()
	file.CreatedAt = r.clock.Now().UTC()
	file.CreatedBy = r.UserId()

	_, err := r.txn.ModelContext(span.Context(), file).
		Insert(file)
	if err != nil {
		return errors.Wrap(err, "failed to create file record")
	}

	span.Data["fileId"] = file.FileId
	span.Status = sentry.SpanStatusOK

	return nil

}
