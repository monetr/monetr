package repository

import (
	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/sirupsen/logrus"
)

type repositoryBase struct {
	userId    ID[User]
	accountId ID[Account]
	txn       pg.DBI
	account   *Account
	kms       secrets.KeyManagement
	clock     clock.Clock
	log       *logrus.Entry
}
