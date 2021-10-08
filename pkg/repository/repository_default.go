package repository

import (
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/pkg/models"
)

type repositoryBase struct {
	userId, accountId uint64
	txn               pg.DBI
	account           *models.Account
}
