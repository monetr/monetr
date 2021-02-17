//+build vault

package repository

import (
	"github.com/go-pg/pg/v10"
	"github.com/hashicorp/vault/api"
)

type repositoryBase struct {
	userId, accountId uint64
	txn               *pg.Tx
	vault             *api.Client
}
