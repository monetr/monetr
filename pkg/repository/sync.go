package repository

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

// GetLastPlaidSync will return the last Plaid sync object for the provided plaid link. If there has not yet been a
// Plaid sync for that link then this will return nil.
func (r *repositoryBase) GetLastPlaidSync(ctx context.Context, plaidLinkId uint64) (*models.PlaidSync, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"plaidLinkId": plaidLinkId,
	}

	var sync models.PlaidSync
	err := r.txn.ModelContext(span.Context(), &sync).
		Where(`"plaid_sync"."plaid_link_id" = ?`, plaidLinkId).
		Order(`timestamp DESC`).
		Limit(1).
		Select(&sync)
	switch err {
	case nil:
		return &sync, nil
	case pg.ErrNoRows:
		// If we didn't receive anything then that means there has not been a sync yet.
		return nil, nil
	default:
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve the latest plaid sync for plaid link")
	}
}

func (r *repositoryBase) RecordPlaidSync(
	ctx context.Context,
	plaidLinkId uint64,
	nextCursor string,
	trigger string,
	added, modified, removed int,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]interface{}{
		"plaidLinkId": plaidLinkId,
	}

	item := models.PlaidSync{
		PlaidLinkID: plaidLinkId,
		Timestamp:   time.Now().UTC(),
		Trigger:     trigger,
		NextCursor:  nextCursor,
		Added:       added,
		Modified:    modified,
		Removed:     removed,
	}

	_, err := r.txn.ModelContext(span.Context(), &item).
		Insert(&item)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to record plaid sync")
	}

	return nil
}
