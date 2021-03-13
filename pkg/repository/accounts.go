package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetAccount() (*models.Account, error) {
	var result models.Account
	err := r.txn.Model(&result).
		Where(`"account"."account_id" = ?`, r.AccountId()).
		Limit(1).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve account")
	}

	return &result, nil
}
