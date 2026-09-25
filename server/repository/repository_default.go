package repository

import (
	"log/slog"

	"github.com/benbjohnson/clock"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/uptrace/bun"
)

type repositoryBase struct {
	userId    ID[User]
	accountId ID[Account]
	txn       bun.IDB
	account   *Account
	kms       secrets.KeyManagement
	clock     clock.Clock
	log       *slog.Logger
}
