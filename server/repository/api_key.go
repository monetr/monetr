package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetApiKeys(
	ctx context.Context,
) ([]models.ApiKey, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]models.ApiKey, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"api_key"."account_id" = ?`, r.AccountId()).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve api keys")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}
