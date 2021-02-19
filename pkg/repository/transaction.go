package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) InsertTransactions(transactions []models.Transaction) error {
	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
	}
	_, err := r.txn.Model(&transactions).Insert(&transactions)
	return errors.Wrap(err, "failed to insert transactions")
}
