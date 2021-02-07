package repository

import (
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

type Repository interface {
	UserId() uint64
	AccountId() uint64
	GetMe() (*models.User, error)
	GetIsSetup() (bool, error)

	CreateLink(link *models.Link) error
	CreateBankAccounts(bankAccounts []models.BankAccount) error

	GetLinks() ([]models.Link, error)
}

type UnauthenticatedRepository interface {
	CreateLogin(email, hashedPassword string, isEnabled bool) (*models.Login, error)
	CreateAccount(timezone *time.Location) (*models.Account, error)
	CreateUser(loginId, accountId uint64, firstName, lastName string) (*models.User, error)
	CreateRegistration(loginId uint64) (*models.Registration, error)

	// VerifyRegistration takes a registrationId and will finalize the registration record. If the registration has
	// already been completed an error is returned.
	VerifyRegistration(registrationId string) (*models.User, error)
}

func NewRepositoryFromSession(userId, accountId uint64, txn *pg.Tx) Repository {
	return &repositoryBase{
		userId:    userId,
		accountId: accountId,
		txn:       txn,
	}
}

func NewUnauthenticatedRepository(txn *pg.Tx) UnauthenticatedRepository {
	return &unauthenticatedRepo{
		txn: txn,
	}
}

var (
	_ Repository = &repositoryBase{}
)

type repositoryBase struct {
	userId, accountId uint64
	txn               *pg.Tx
}

func (r *repositoryBase) UserId() uint64 {
	return r.userId
}

func (r *repositoryBase) AccountId() uint64 {
	return r.accountId
}

func (r *repositoryBase) GetMe() (*models.User, error) {
	var user models.User
	err := r.txn.Model(&user).
		Relation("Login").
		Relation("Login.EmailVerifications").
		Relation("Login.PhoneVerifications").
		Relation("Account").
		Where(`"user"."user_id" = ? AND "user"."account_id" = ?`, r.userId, r.accountId).
		Limit(1).
		Select(&user)
	switch err {
	case pg.ErrNoRows:
		return nil, errors.Errorf("user does not exist")
	default:
		return nil, errors.Wrapf(err, "failed to retrieve user")
	case nil:
		break
	}

	return &user, nil
}

func (r *repositoryBase) GetIsSetup() (bool, error) {
	return r.txn.Model(&models.Link{}).
		Where(`"link"."account_id" = ?`, r.accountId).
		Exists()
}

func (r *repositoryBase) GetLinks() ([]models.Link, error) {
	var result []models.Link
	err := r.txn.Model(&result).
		Where(`"link"."account_id" = ?`, r.accountId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve links")
	}

	return result, nil
}

func (r *repositoryBase) CreateLink(link *models.Link) error {
	_, err := r.txn.Model(link).Insert(link)
	return errors.Wrap(err, "failed to insert link")
}

func (r *repositoryBase) CreateBankAccounts(bankAccounts []models.BankAccount) error {
	_, err := r.txn.Model(&bankAccounts).Insert(&bankAccounts)
	return errors.Wrap(err, "failed to insert bank accounts")
}

func (r *repositoryBase) GetBankAccounts() ([]models.BankAccount, error) {
	return nil, nil
}
