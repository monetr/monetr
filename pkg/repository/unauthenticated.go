package repository

import (
	"context"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/hash"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

var (
	_ UnauthenticatedRepository = &unauthenticatedRepo{}
)

var (
	ErrEmailAlreadyExists = errors.New("a login with the same email already exists")
)

type unauthenticatedRepo struct {
	txn pg.DBI
}

func (u *unauthenticatedRepo) GetLoginForChallenge(ctx context.Context, email string) (*models.LoginWithVerifier, error) {
	span := sentry.StartSpan(ctx, "GetLoginForChallenge")
	defer span.Finish()

	email = strings.ToLower(email)
	var login models.LoginWithVerifier
	err := u.txn.ModelContext(span.Context(), &login).
		Where(`"email" = ?`, email).
		Limit(1).
		Select(&login)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve login details")
	}

	return &login, nil
}

func (u *unauthenticatedRepo) GetLoginById(ctx context.Context, loginId uint64) (*models.Login, error) {
	return nil, nil
}

func (u *unauthenticatedRepo) CreateSecureLogin(ctx context.Context, newLogin *models.LoginWithVerifier) error {
	span := sentry.StartSpan(ctx, "CreateSecureLogin")
	defer span.Finish()

	// Just to make sure, clean up these fields.
	newLogin.Email = strings.ToLower(newLogin.Email)
	newLogin.LoginId = 0

	count, err := u.txn.ModelContext(span.Context(), newLogin).
		Where(`"email" = ?`, newLogin.Email).
		Count()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to verify if email is unique")
	}

	if count != 0 {
		span.Status = sentry.SpanStatusInvalidArgument
		return errors.WithStack(ErrEmailAlreadyExists)
	}

	_, err = u.txn.ModelContext(span.Context(), newLogin).Insert(newLogin)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
	}

	return errors.Wrap(err, "failed to create new login")
}

func (u *unauthenticatedRepo) CreateLogin(
	ctx context.Context,
	email, hashedPassword string, firstName, lastName string,
) (*models.Login, error) {
	span := sentry.StartSpan(ctx, "CreateLogin")
	defer span.Finish()

	login := &models.LoginWithHash{
		Login: models.Login{
			Email:           strings.ToLower(email),
			FirstName:       firstName,
			LastName:        lastName,
			IsEnabled:       true,
			IsEmailVerified: EmailNotVerified, // Always insert false.
		},
		PasswordHash: hashedPassword,
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

func (u *unauthenticatedRepo) CreateAccountV2(ctx context.Context, account *models.Account) error {
	span := sentry.StartSpan(ctx, "CreateAccount")
	defer span.Finish()

	// Make sure that the Id is not being specified by the caller of this method. Will ensure that the caller cannot
	// specify the accountId and overwrite something.
	account.AccountId = 0

	_, err := u.txn.ModelContext(span.Context(), account).Insert(account)
	return errors.Wrap(err, "failed to create account")
}

func (u *unauthenticatedRepo) CreateUser(ctx context.Context, loginId, accountId uint64, user *models.User) error {
	span := sentry.StartSpan(ctx, "CreateAccount")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"loginId":   loginId,
		"accountId": accountId,
	}

	user.UserId = 0
	user.AccountId = accountId
	user.LoginId = loginId

	if _, err := u.txn.ModelContext(span.Context(), user).Insert(user); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create user")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (u *unauthenticatedRepo) GetLoginForEmail(ctx context.Context, emailAddress string) (*models.Login, error) {
	return getLoginForEmail(ctx, u.txn, emailAddress)
}

func (u *unauthenticatedRepo) GetLinksForItem(ctx context.Context, itemId string) (*models.Link, error) {
	span := sentry.StartSpan(ctx, "GetLinksForItem")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"itemId": itemId,
	}

	var link models.Link
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

func (u *unauthenticatedRepo) ValidateBetaCode(ctx context.Context, betaCode string) (*models.Beta, error) {
	span := sentry.StartSpan(ctx, "Validate Beta Code")
	defer span.Finish()
	var beta models.Beta
	hashedCode := hash.HashPassword(strings.ToLower(betaCode), betaCode)
	err := u.txn.ModelContext(span.Context(), &beta).
		Where(`"beta"."code_hash" = ?`, hashedCode).
		Where(`"beta"."used_by_user_id" IS NULL`).
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

func (u *unauthenticatedRepo) UseBetaCode(ctx context.Context, betaId, usedBy uint64) error {
	span := sentry.StartSpan(ctx, "Use Beta Code")
	defer span.Finish()
	result, err := u.txn.ModelContext(span.Context(), &models.Beta{}).
		Set(`"used_by_user_id" = ?`, usedBy).
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

func (u *unauthenticatedRepo) ResetPassword(ctx context.Context, loginId uint64, hashedPassword string) error {
	span := sentry.StartSpan(ctx, "ResetPassword")
	defer span.Finish()

	result, err := u.txn.ModelContext(span.Context(), &models.LoginWithHash{}).
		Set(`"password_hash" = ?`, hashedPassword).
		Set(`"password_reset_at" = ?`, time.Now()).
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
