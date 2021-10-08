package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) UpdateUser(ctx context.Context, user *models.User) error {
	span := sentry.StartSpan(ctx, "UpdateUser")
	defer span.Finish()

	user.UserId = r.UserId()
	user.AccountId = r.AccountId()

	result, err := r.txn.ModelContext(span.Context(), user).
		WherePK().
		Update(user)
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if affected := result.RowsAffected(); affected != 1 {
		return errors.Errorf("invalid number of user(s) updated; expected: 1 updated: %d", affected)
	}

	return nil
}
