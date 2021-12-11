package repository

import (
	"github.com/monetr/monetr/pkg/models"
	"github.com/uptrace/bun"
)

type repositoryBase struct {
	userId, accountId uint64
	db                bun.IDB
	account           *models.Account
}
