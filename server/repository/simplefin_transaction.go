package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateSimpleFinTransactions(
	ctx context.Context,
	transactions []SimpleFinTransaction,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())

	now := r.clock.Now().UTC()
	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
		transactions[i].CreatedAt = now
	}

	if _, err := r.txn.NewInsert().
		Model(&transactions).
		Returning("*").
		Exec(span.Context()); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create SimpleFIN transactions")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}
