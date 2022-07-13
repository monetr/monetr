package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetSettings(ctx context.Context) (*models.Settings, error) {
	span := sentry.StartSpan(ctx, "GetSettings")
	defer span.Finish()

	result := models.Settings{
		AccountId: r.AccountId(),
		MaxSafeToSpend: struct {
			Enabled bool  `json:"enabled"`
			Maximum int64 `json:"maximum"`
		}{
			Enabled: false,
			Maximum: 0,
		},
	}
	inserted, err := r.txn.ModelContext(span.Context(), &result).
		Where(`"settings"."account_id" = ?`, r.AccountId()).
		Limit(1).
		SelectOrInsert(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve account settings")
	}
	if inserted {
		crumbs.Debug(span.Context(), "settings were not present, default settings were created", map[string]interface{}{
			"settings": result,
		})
	}

	return &result, nil
}
