package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetAccount(ctx context.Context) (*models.Account, error) {
	if r.account != nil {
		return r.account, nil
	}

	span := sentry.StartSpan(ctx, "GetAccount")
	defer span.Finish()

	var account models.Account
	err := r.db.NewSelect().Model(&account).
		Where(`account.account_id = ?`, r.AccountId()).
		Limit(1).
		Scan(span.Context(), &account)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve account")
	}

	span.Status = sentry.SpanStatusOK

	r.account = &account

	return r.account, nil
}
