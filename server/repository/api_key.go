package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetApiKeyById(
	ctx context.Context,
	id models.ID[models.ApiKey],
) (*models.ApiKey, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var result models.ApiKey
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("CreatedByUser").
		Relation("CreatedByUser.Login").
		Where(`"api_key"."api_key_id" = ?`, id).
		Where(`"api_key"."account_id" = ?`, r.AccountId()).
		Limit(1).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve api key")
	}

	return &result, nil
}

func (r *repositoryBase) GetApiKeys(
	ctx context.Context,
) ([]models.ApiKey, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]models.ApiKey, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"api_key"."account_id" = ?`, r.AccountId()).
		Where(`"api_key"."deleted_at" IS NULL`).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve api keys")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) CreateApiKey(
	ctx context.Context,
	key *models.ApiKey,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	key.AccountId = r.AccountId()
	key.CreatedBy = r.UserId()

	if _, err := r.txn.ModelContext(
		span.Context(),
		key,
	).Insert(key); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create api key")
	}
	return nil
}

func (r *repositoryBase) DeleteApiKey(
	ctx context.Context,
	id models.ID[models.ApiKey],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result, err := r.txn.ModelContext(span.Context(), new(models.ApiKey)).
		Set(`"deleted_at" = ?`, r.clock.Now()).
		Where(`"api_key"."api_key_id" = ?`, id).
		Where(`"api_key"."account_id" = ?`, r.AccountId()).
		Where(`"api_key"."deleted_at" IS NULL`).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to delete api key")
	} else if result.RowsAffected() == 0 {
		// Return this error so our upstream wrapPgError handles it properly
		return errors.Wrap(pg.ErrNoRows, "invalid api key specified")
	}

	return nil
}
