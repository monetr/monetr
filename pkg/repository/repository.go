package repository

import (
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
)

type Repository interface {
	UserId() uint64
	AccountId() uint64
}

type UnauthenticatedRepository interface {
	CreateLogin(email, hashedPassword string) (*models.Login, error)
	CreateAccount(timezone *time.Location) (*models.Account, error)
	CreateUser(loginId, accountId uint64, firstName, lastName string) (*models.User, error)
}

func NewRepositoryFromSession(userId, accountId uint64, txn *pg.Tx) Repository {
	return nil
}

func NewUnauthenticatedRepository(txn *pg.Tx) UnauthenticatedRepository {
	return &unauthenticatedRepo{
		txn: txn,
	}
}
