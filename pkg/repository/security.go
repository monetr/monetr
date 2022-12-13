package repository

import (
	"context"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/consts"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	// Deprecated: use requiresPasswordChange instead ErrRequirePasswordChange is an error that is returned when a user's
	// password must be updated. This is more or less future planning, as I'd like to implement a new method of storing
	// passwords and need to make sure a path to make user's update their passwords is in place. This error should not be
	// considered a "failure" state.
	ErrRequirePasswordChange = errors.New("password must be updated")
	// ErrInvalidCredentials is returned when the provided email and/or password does not match the current password for
	// an existing login.
	ErrInvalidCredentials = errors.New("invalid credentials provided")
)

type SecurityRepository interface {
	// Login recieves the user's provided email address as well as password. This password is then hashed using bcrypt and
	// is then compared to the stored hash for that login. If the provided password is equivalent to the stored hash then
	// this function will return the login model for the credentials provided. As well as a boolean indicating whether or
	// not the user must change their password at this time. A user MUST NOT be given credentials if their password
	// requires changing. If the provided credentials are invalid then ErrInvalidCredentials is returned.
	// NOTE: ErrRequirePasswordChange is no longer returned.
	Login(ctx context.Context, email, password string) (_ *models.Login, requiresPasswordChange bool, err error)

	// ChangePassword accepts a login ID and the old hashed password and the new hashed password. The two passwords
	// should be hashed from the user's input. Specifically, you should not retrieve the "oldHashedPassword" from the
	// database and then use it as input for this method. This way the function will only succeed if the provided input
	// is 100% valid. This method will return true if the oldHashedPassword is correct and the update succeeds, it will
	// return false if the oldHashedPassword is incorrect and/or if the update fail. If the oldHashedPassword provided
	// is not valid for the login ID, then ErrInvalidCredentials will be returned.
	ChangePassword(ctx context.Context, loginId uint64, oldHashedPassword, newHashedPassword string) error
}

var (
	_ SecurityRepository = &baseSecurityRepository{}
)

type baseSecurityRepository struct {
	db pg.DBI
}

func NewSecurityRepository(db pg.DBI) SecurityRepository {
	return &baseSecurityRepository{
		db: db,
	}
}

func (b *baseSecurityRepository) Login(ctx context.Context, email, password string) (*models.Login, bool, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	requiresPasswordChange := false

	var login models.LoginWithHash
	err := b.db.ModelContext(span.Context(), &login).
		Relation("Users").
		Relation("Users.Account").
		Where(`"login_with_hash"."email" = ?`, strings.ToLower(email)).
		Limit(1).
		Select(&login)
	switch err {
	case nil:
	case pg.ErrNoRows:
		span.Status = sentry.SpanStatusNotFound
		return nil, requiresPasswordChange, errors.WithStack(ErrInvalidCredentials)
	default:
		span.Status = sentry.SpanStatusInternalError
		return nil, requiresPasswordChange, crumbs.WrapError(span.Context(), err, "failed to verify credentials")
	}

	if err = bcrypt.CompareHashAndPassword(login.Crypt, []byte(password)); err != nil {
		return nil, requiresPasswordChange, errors.WithStack(ErrInvalidCredentials)
	}

	span.Status = sentry.SpanStatusOK
	return &login.Login, requiresPasswordChange, nil
}

func (b *baseSecurityRepository) ChangePassword(ctx context.Context, loginId uint64, oldPassword, newPassword string) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var login models.LoginWithHash
	err := b.db.ModelContext(span.Context(), &login).
		Where(`"login_id" = ?`, loginId).
		Limit(1).
		Select(&login)
	if err != nil {
		return crumbs.WrapError(span.Context(), err, "failed to find login record to change password")
	}

	if err = bcrypt.CompareHashAndPassword(login.Crypt, []byte(oldPassword)); err != nil {
		return errors.WithStack(ErrInvalidCredentials)
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), consts.BcryptCost)
	if err != nil {
		return crumbs.WrapError(span.Context(), err, "failed to encrypt new password for change")
	}

	_, err = b.db.ModelContext(span.Context(), &login).
		Set(`"crypt" = ?`, newPasswordHash).
		Where(`"login_id" = ?`, loginId).
		Update(&login)
	if err != nil {
		return crumbs.WrapError(span.Context(), err, "failed to update password")
	}

	return nil
}
