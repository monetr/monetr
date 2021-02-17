//+build !vault

package repository

import (
	"github.com/go-pg/pg/v10"
)

type repositoryBase struct {
	userId, accountId uint64
	txn               *pg.Tx
}
