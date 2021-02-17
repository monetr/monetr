package repository

import (
	"github.com/go-pg/pg/v10"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

type JobRepository interface {
	GetBankAccountsToSync() ([]models.BankAccount, error)
}

type jobRepository struct {
	txn *pg.Tx
}

func (j *jobRepository) GetBankAccountsToSync() ([]models.BankAccount, error) {
	var result []models.BankAccount
	err := j.txn.Model(&result).
		Relation("Link").
		Relation("Link.PlaidLink").
		Where(`"link"."link_type" = ?`, models.PlaidLinkType).
		Select(&result)
	return result, errors.Wrap(err, "failed to retrieve bank accounts to sync")
}
