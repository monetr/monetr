package repository

import (
	"context"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/hash"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
)

var (
	_ UnauthenticatedRepository = &unauthenticatedRepo{}
)

type unauthenticatedRepo struct {
	txn *pg.Tx
}

func (u *unauthenticatedRepo) CreateLogin(
	email, hashedPassword string, firstName, lastName string, isEnabled bool,
) (*models.Login, error) {
	login := &models.Login{
		Email:        strings.ToLower(email),
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		IsEnabled:    isEnabled,
	}
	count, err := u.txn.Model(login).
		Where(`"login"."email" = ?`, email).
		Count()
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify if email is unique")
	}

	if count != 0 {
		return nil, errors.Errorf("a login with the same email already exists")
	}

	_, err = u.txn.Model(login).Insert(login)
	return login, errors.Wrap(err, "failed to create login")
}

func (u *unauthenticatedRepo) CreateAccount(timezone *time.Location) (*models.Account, error) {
	account := &models.Account{
		Timezone: timezone.String(),
	}
	_, err := u.txn.Model(account).Insert(account)
	return account, errors.Wrap(err, "failed to create account")
}

func (u *unauthenticatedRepo) CreateAccountV2(ctx context.Context, account *models.Account) error {
	span := sentry.StartSpan(ctx, "Create Account")
	defer span.Finish()

	// Make sure that the Id is not being specified by the caller of this method. Will ensure that the caller cannot
	// specify the accountId and overwrite something.
	account.AccountId = 0

	_, err := u.txn.ModelContext(span.Context(), account).Insert(account)
	return errors.Wrap(err, "failed to create account")
}

func (u *unauthenticatedRepo) CreateUser(loginId, accountId uint64, user *models.User) error {
	user.UserId = 0
	user.AccountId = accountId
	user.LoginId = loginId

	if _, err := u.txn.Model(user).Insert(user); err != nil {
		return errors.Wrap(err, "failed to create user")
	}

	return nil
}

func (u *unauthenticatedRepo) VerifyRegistration(registrationId string) (*models.User, error) {
	panic("not implemented")
}

func (u *unauthenticatedRepo) GetLinksForItem(itemId string) (*models.Link, error) {
	var link models.Link
	err := u.txn.Model(&link).
		Relation("PlaidLink").
		Where(`"plaid_link"."item_id" = ?`, itemId).
		Limit(1).
		Select(&link)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve plaid link")
	}

	if link.PlaidLink == nil {
		return nil, errors.Errorf("failed to retrieve link for item id")
	}

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
	span := sentry.StartSpan(ctx, "Validate Beta Code")
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
