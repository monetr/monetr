package repository

import (
	"context"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/consts"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/hash"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrRequirePasswordChange is an error that is returned when a user's password must be updated. This is more or
	// less future planning, as I'd like to implement a new method of storing passwords and need to make sure a path to
	// make user's update their passwords is in place. This error should not be considered a "failure" state.
	ErrRequirePasswordChange = errors.New("password must be updated")
	// ErrInvalidCredentials is returned when the provided email and/or password does not match the current password for
	// an existing login.
	ErrInvalidCredentials = errors.New("invalid credentials provided")
)

type SecurityRepository interface {
	// LoginOld receives the user's provided email address, as well as a hashed version of their password. It will then
	// verify that a login with that email and password does exist and return that login. This method can return the
	// error ErrRequirePasswordChange. This method can also return ErrInvalidCredentials. If the login is not nil then
	// this function call should not be treated as a failure. The caller should check for one of the two aforementioned
	// errors where this function is called and handle them appropriately.
	// DEPRECATED! Use Login instead.
	LoginOld(ctx context.Context, email, hashedPassword string) (*models.Login, error)

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

func (b *baseSecurityRepository) LoginOld(ctx context.Context, email, hashedPassword string) (*models.Login, error) {
	span := sentry.StartSpan(ctx, "LoginOld")
	defer span.Finish()

	// This shouldn't be necessary, but better safe than sorry.
	email = strings.TrimSpace(strings.ToLower(email))

	var login models.Login
	err := b.db.ModelContext(span.Context(), &login).
		Relation("Users").
		Relation("Users.Account").
		Where(`"login"."email" = ?`, email).
		Where(`"login"."password_hash" = ?`, hashedPassword).
		Limit(1).
		Select(&login)
	switch err {
	case pg.ErrNoRows:
		span.Status = sentry.SpanStatusNotFound
		return nil, errors.WithStack(ErrInvalidCredentials)
	case nil:
		span.Status = sentry.SpanStatusOK
		return &login, nil
	default:
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(span.Context(), err, "failed to verify credentials")
	}
}

func (b *baseSecurityRepository) Login(ctx context.Context, email, password string) (*models.Login, bool, error) {
	span := sentry.StartSpan(ctx, "Login")
	defer span.Finish()

	requiresPasswordChange := false

	// Used for backwards compatability.
	hashedPassword := hash.HashPassword(email, password)

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

	switch {
	case login.Crypt != nil:
		if err = bcrypt.CompareHashAndPassword(login.Crypt, []byte(password)); err != nil {
			return nil, requiresPasswordChange, errors.WithStack(ErrInvalidCredentials)
		}
	case login.PasswordHash != nil:
		if *login.PasswordHash != hashedPassword {
			return nil, requiresPasswordChange, errors.WithStack(ErrInvalidCredentials)
		}
		requiresPasswordChange = true
	default:
		span.Status = sentry.SpanStatusDataLoss
		crumbs.IndicateBug(
			span.Context(),
			"Login is missing bcrypt and legacy password hash, cannot authenticate",
			map[string]interface{}{
				"loginId": login.LoginId,
			})
		return nil, requiresPasswordChange, errors.Errorf("failed to verify credentials")
	}

	span.Status = sentry.SpanStatusOK
	return &login.Login, requiresPasswordChange, nil
}

func (b *baseSecurityRepository) ChangePassword(ctx context.Context, loginId uint64, oldPassword, newPassword string) error {
	span := sentry.StartSpan(ctx, "ChangePassword")
	defer span.Finish()

	var login models.LoginWithHash
	err := b.db.ModelContext(span.Context(), &login).
		Where(`"login_id" = ?`, loginId).
		Limit(1).
		Select(&login)
	if err != nil {
		return crumbs.WrapError(span.Context(), err, "failed to find login record to change password")
	}

	switch {
	case login.Crypt != nil:
		if err = bcrypt.CompareHashAndPassword(login.Crypt, []byte(oldPassword)); err != nil {
			return errors.WithStack(ErrInvalidCredentials)
		}
	case login.PasswordHash != nil:
		hashedPassword := hash.HashPassword(login.Email, newPassword)
		if *login.PasswordHash != hashedPassword {
			return errors.WithStack(ErrInvalidCredentials)
		}
	default:
		span.Status = sentry.SpanStatusDataLoss
		crumbs.IndicateBug(
			span.Context(),
			"Login is missing bcrypt and legacy password hash, cannot authenticate",
			map[string]interface{}{
				"loginId": login.LoginId,
			})
		return errors.Errorf("failed to verify credentials")
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
