package repository

import (
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetrapp/rest-api/pkg/models"
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