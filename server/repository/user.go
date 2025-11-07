package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) UpdateUser(ctx context.Context, user *User) error {
	span := crumbs.StartFnTrace(ctx)
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

func (r *repositoryBase) GetMe(ctx context.Context) (*User, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId": r.AccountId(),
		"userId":    r.UserId(),
	}

	var user User
	err := r.txn.ModelContext(span.Context(), &user).
		Relation("Login").
		Relation("Account").
		Where(`"user"."user_id" = ? AND "user"."account_id" = ?`, r.userId, r.accountId).
		Limit(1).
		Select(&user)
	switch err {
	case pg.ErrNoRows:
		span.Status = sentry.SpanStatusNotFound
		return nil, errors.Errorf("user does not exist")
	case nil:
		break
	default:
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrapf(err, "failed to retrieve user")
	}

	span.Status = sentry.SpanStatusOK

	return &user, nil
}

// GetAccountOwner will return a User object for the currently authenticated
// account, as well as the Login and Account sub object for that user. If one is
// not found then an error is returned.
func (r *repositoryBase) GetAccountOwner(ctx context.Context) (*User, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId": r.AccountId(),
	}

	var user User
	err := r.txn.ModelContext(span.Context(), &user).
		Relation("Login").
		Relation("Account").
		Where(`"user"."account_id" = ?`, r.AccountId()).
		Where(`"user"."role" = ?`, UserRoleOwner).
		Limit(1).
		Select(&user)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find account owner")
	}

	return &user, nil
}
