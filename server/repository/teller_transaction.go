package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateTellerTransaction(ctx context.Context, transaction *models.TellerTransaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transaction.AccountId = r.AccountId()
	transaction.CreatedAt = r.clock.Now().UTC()
	transaction.UpdatedAt = r.clock.Now().UTC()

	_, err := r.txn.ModelContext(span.Context(), transaction).Insert(transaction)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create teller transaction")
	}

	return nil
}

func (r *repositoryBase) DeleteTellerTransaction(ctx context.Context, transaction *models.TellerTransaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	transaction.AccountId = r.AccountId()
	transaction.DeletedAt = myownsanity.TimeP(r.clock.Now().UTC())

	_, err := r.txn.ModelContext(span.Context(), transaction).
		Set(`"deleted_at" = ?`, *transaction.DeletedAt).
		WherePK().
		Update()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to soft-delete teller transaction")
	}

	return nil
}
