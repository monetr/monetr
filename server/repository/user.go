package repository

import (
	"context"
	"database/sql"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) UpdateUser(ctx context.Context, user *User) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	user.UserId = r.UserId()
	user.AccountId = r.AccountId()

	result, err := r.txn.NewUpdate().
		Model(user).
		WherePK().
		Returning("*").
		Exec(span.Context())
	if err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if affected, _ := result.RowsAffected(); affected != 1 {
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
	err := r.txn.NewSelect().
		Model(&user).
		Relation("Login").
		Relation("Account").
		Where(`"user"."user_id" = ? AND "user"."account_id" = ?`, r.userId, r.accountId).
		Limit(1).
		Scan(span.Context())
	switch err {
	case sql.ErrNoRows:
		span.Status = sentry.SpanStatusNotFound
		return nil, errors.Errorf("user does not exist")
	case nil:
	default:
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrapf(err, "failed to retrieve user")
	}

	span.Status = sentry.SpanStatusOK

	return &user, nil
}

func (r *repositoryBase) GetUserById(
	ctx context.Context,
	id ID[User],
) (*User, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var user User
	err := r.txn.NewSelect().
		Model(&user).
		Relation("Login").
		Relation("Account").
		Where(`"user"."account_id" = ?`, r.AccountId()).
		Where(`"user"."user_id" = ?`, id).
		Limit(1).
		Scan(span.Context())
	switch err {
	case sql.ErrNoRows:
		span.Status = sentry.SpanStatusNotFound
		// Keep sql.ErrNoRows as the cause so the controller can translate this
		// into a 404 instead of a 500.
		return nil, errors.Wrap(err, "user does not exist")
	case nil:
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
	err := r.txn.NewSelect().
		Model(&user).
		Relation("Login").
		Relation("Account").
		Where(`"user"."account_id" = ?`, r.AccountId()).
		Where(`"user"."role" = ?`, UserRoleOwner).
		Limit(1).
		Scan(span.Context())
	if err != nil {
		return nil, errors.Wrap(err, "failed to find account owner")
	}

	return &user, nil
}
