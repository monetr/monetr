package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

// GetLastPlaidSync will return the last Plaid sync object for the provided plaid link. If there has not yet been a
// Plaid sync for that link then this will return nil.
func (r *repositoryBase) GetLastPlaidSync(ctx context.Context, plaidLinkId ID[PlaidLink]) (*PlaidSync, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]any{
		"plaidLinkId": plaidLinkId,
	}

	var sync PlaidSync
	err := r.txn.ModelContext(span.Context(), &sync).
		Where(`"plaid_sync"."plaid_link_id" = ?`, plaidLinkId).
		Where(`"plaid_sync"."account_id" = ?`, r.AccountId()).
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
	plaidLinkId ID[PlaidLink],
	nextCursor string,
	trigger string,
	added, modified, removed int,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.Data = map[string]any{
		"plaidLinkId": plaidLinkId,
	}

	item := PlaidSync{
		PlaidLinkId: plaidLinkId,
		AccountId:   r.AccountId(),
		Timestamp:   r.clock.Now().UTC(),
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
