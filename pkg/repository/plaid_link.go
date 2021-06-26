//+build !vault

package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreatePlaidLink(link *models.PlaidLink) error {
	_, err := r.txn.Model(link).Insert(link)
	return errors.Wrap(err, "failed to create plaid link")
}

func (r *repositoryBase) UpdatePlaidLink(ctx context.Context, link *models.PlaidLink) error {
	span := sentry.StartSpan(ctx, "UpdatePlaidLink")
	defer span.Finish()
	if span.Data == nil {
		span.Data = map[string]interface{}{}
	}

	span.SetTag("accountId", r.AccountIdStr())
	_, err := r.txn.ModelContext(span.Context(), link).WherePK().Update(link)
	return errors.Wrap(err, "failed to update Plaid link")
}