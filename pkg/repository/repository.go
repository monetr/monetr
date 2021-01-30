package repository

import (
	"github.com/go-pg/pg/v10"
)

type Repository interface {
	UserId() uint64
	AccountId() uint64
}

func NewRepositoryFromSession(userId, accountId uint64, txn *pg.Tx) Repository {
	return nil
}
