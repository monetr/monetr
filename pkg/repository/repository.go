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
}

type UnauthenticatedRepository interface {
	CreateLogin(email, hashedPassword string) (*models.Login, error)
	CreateAccount(timezone *time.Location) (*models.Account, error)
	CreateUser(loginId, accountId uint64, firstName, lastName string) (*models.User, error)
	CreateRegistration(loginId uint64) (*models.Registration, error)
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
