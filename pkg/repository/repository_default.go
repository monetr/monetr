//+build !vault

package repository

import (
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
)

type repositoryBase struct {
	userId, accountId uint64
	txn               pg.DBI
	account           *models.Account
}
