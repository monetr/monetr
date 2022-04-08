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
	err := r.txn.ModelContext(span.Context(), &account).
		Where(`"account"."account_id" = ?`, r.AccountId()).
		Limit(1).
		Select(&account)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve account")
	}

	span.Status = sentry.SpanStatusOK

	r.account = &account

	return r.account, nil
}

// DeleteAccount removes all of the records from the database related to the current account. This action cannot be
// undone. Any Plaid links should be removed BEFORE calling this function.
func (r *repositoryBase) DeleteAccount(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "DeleteAccount")
	defer span.Finish()

	dataTypes := []interface{}{
		&models.Transaction{},
		&models.Spending{},
		&models.FundingSchedule{},
		&models.BankAccount{},
		&models.Link{},
		&models.User{},
		&models.Account{},
	}

	for _, model := range dataTypes {
		_, err := r.txn.ModelContext(span.Context(), model).
			Where(`"account_id" = ?`, r.AccountId()).
			ForceDelete()
		if err != nil {
			return errors.Wrapf(err, "failed to delete %T for account", model)
		}
	}

	return nil
}
