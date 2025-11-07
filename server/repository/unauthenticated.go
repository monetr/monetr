package repository

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/consts"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("a login with the same email already exists")
)

var (
	_ UnauthenticatedRepository = &unauthenticatedRepo{}
)

type EmailVerification = bool

const (
	EmailVerified    EmailVerification = true
	EmailNotVerified EmailVerification = false
)

type unauthenticatedRepo struct {
	txn   pg.DBI
	clock clock.Clock
}

func (u *unauthenticatedRepo) CreateLogin(
	ctx context.Context,
	email, password string, firstName, lastName string,
) (*Login, error) {
	span := sentry.StartSpan(ctx, "CreateLogin")
	defer span.Finish()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), consts.BcryptCost)
	if err != nil {
		return nil, errors.Wrap(err, "failed to bcrypt provided password")
	}

	login := &LoginWithHash{
		Login: Login{
			Email:           strings.ToLower(email),
			FirstName:       firstName,
			LastName:        lastName,
			IsEnabled:       true,
			IsEmailVerified: EmailNotVerified, // Always insert false.
		},
		Crypt: hashedPassword,
	}
	count, err := u.txn.ModelContext(span.Context(), login).
		Where(`"email" = ?`, email).
		Count()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to verify if email is unique")
	}

	if count != 0 {
		span.Status = sentry.SpanStatusInvalidArgument
		return nil, errors.WithStack(ErrEmailAlreadyExists)
	}

	_, err = u.txn.ModelContext(span.Context(), login).Insert(login)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
	}

	return &login.Login, errors.Wrap(err, "failed to create login")
}

func (u *unauthenticatedRepo) CreateAccountV2(ctx context.Context, account *Account) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Make sure that the Id is not being specified by the caller of this method. Will ensure that the caller cannot
	// specify the accountId and overwrite something.
	account.AccountId = ""
	// TODO Should this be in the timezone of the account? Time is time right, now should be the same no matter what?
	account.CreatedAt = time.Now()

	_, err := u.txn.ModelContext(span.Context(), account).Insert(account)
	return errors.Wrap(err, "failed to create account")
}

func (u *unauthenticatedRepo) CreateUser(
	ctx context.Context,
	user *User,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"loginId":   user.LoginId,
		"accountId": user.AccountId,
	}

	user.UserId = ""

	if _, err := u.txn.ModelContext(span.Context(), user).Insert(user); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create user")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (u *unauthenticatedRepo) GetLoginForEmail(ctx context.Context, emailAddress string) (*Login, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var login Login
	err := u.txn.ModelContext(span.Context(), &login).
		Where(`"login"."email" = ?`, strings.ToLower(emailAddress)). // Only for a login with this email.
		Limit(1).
		Select(&login)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve login by email")
	}

	return &login, nil
}

func (u *unauthenticatedRepo) SetEmailVerified(ctx context.Context, emailAddress string) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var login Login
	result, err := u.txn.ModelContext(span.Context(), &login).
		Set(`"is_email_verified" = ?`, EmailVerified).               // Change the verification to true.
		Set(`"email_verified_at" = ?`, u.clock.Now().UTC()).         // Set the verified at time to now.
		Where(`"login"."email" = ?`, strings.ToLower(emailAddress)). // Only for a login with this email.
		Where(`"login"."is_enabled" = ?`, true).                     // Only if the login is actually enabled.
		Where(`"login"."is_email_verified" = ?`, EmailNotVerified).  // And only if the login is not already verified.
		Limit(1).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to verify email")
	}

	if result.RowsAffected() != 1 {
		return errors.New("email cannot be verified")
	}

	return nil
}

func (u *unauthenticatedRepo) GetLinksForItem(ctx context.Context, itemId string) (*Link, error) {
	span := sentry.StartSpan(ctx, "GetLinksForItem")
	defer span.Finish()

	span.Data = map[string]any{
		"itemId": itemId,
	}

	var link Link
	err := u.txn.ModelContext(span.Context(), &link).
		Relation("PlaidLink").
		Where(`"plaid_link"."item_id" = ?`, itemId).
		Limit(1).
		Select(&link)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve plaid link")
	}

	if link.PlaidLink == nil {
		span.Status = sentry.SpanStatusNotFound
		return nil, errors.Errorf("failed to retrieve link for item id")
	}

	span.Status = sentry.SpanStatusOK

	return &link, nil
}

func (u *unauthenticatedRepo) ValidateBetaCode(ctx context.Context, betaCode string) (*Beta, error) {
	span := sentry.StartSpan(ctx, "Validate Beta Code")
	defer span.Finish()
	var beta Beta

	hash := sha256.New()
	hash.Write([]byte(strings.ToLower(betaCode)))
	hashedCode := fmt.Sprintf("%X", hash.Sum(nil))

	err := u.txn.ModelContext(span.Context(), &beta).
		Where(`"beta"."code_hash" = ?`, hashedCode).
		Where(`"beta"."used_by" IS NULL`).
		Limit(1).
		Select(&beta)
	if err != nil {
		span.Status = sentry.SpanStatusNotFound
		return nil, errors.Wrap(err, "failed to validate beta code")
	}

	if time.Now().After(beta.ExpiresAt) {
		span.Status = sentry.SpanStatusResourceExhausted
		return nil, errors.Errorf("beta code is expired")
	}

	span.Status = sentry.SpanStatusOK

	return &beta, nil
}

func (u *unauthenticatedRepo) UseBetaCode(ctx context.Context, betaId ID[Beta], usedBy ID[User]) error {
	span := sentry.StartSpan(ctx, "Use Beta Code")
	defer span.Finish()
	result, err := u.txn.ModelContext(span.Context(), &Beta{}).
		Set(`"used_by" = ?`, usedBy).
		Where(`"beta"."beta_id" = ?`, betaId).
		Update()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to use beta code")
	}

	if result.RowsAffected() != 1 {
		span.Status = sentry.SpanStatusInvalidArgument
		return errors.Errorf("invalid number of beta codes used: %d", result.RowsAffected())
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (u *unauthenticatedRepo) ResetPassword(ctx context.Context, loginId ID[Login], password string) error {
	span := sentry.StartSpan(ctx, "ResetPassword")
	defer span.Finish()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), consts.BcryptCost)
	if err != nil {
		return crumbs.WrapError(span.Context(), err, "failed to encrypt provided password for reset")
	}

	result, err := u.txn.ModelContext(span.Context(), &LoginWithHash{}).
		Set(`"crypt" = ?`, hashedPassword).
		Set(`"password_reset_at" = ?`, u.clock.Now()).
		Where(`"login_with_hash"."login_id" = ?`, loginId).
		Update()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to reset login password")
	}

	if result.RowsAffected() != 1 {
		span.Status = sentry.SpanStatusNotFound
		return errors.Errorf("no logins were updated")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}
