package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) CreateBankAccounts(bankAccounts ...models.BankAccount) error {
	for i := range bankAccounts {
		bankAccounts[i].BankAccountId = 0
		bankAccounts[i].AccountId = r.AccountId()
	}
	_, err := r.txn.Model(&bankAccounts).Insert(&bankAccounts)
	return errors.Wrap(err, "failed to insert bank accounts")
}
