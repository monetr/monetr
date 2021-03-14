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

func (r *repositoryBase) GetBankAccountsByLinkId(linkId uint64) ([]models.BankAccount, error) {
	var result []models.BankAccount
	err := r.txn.Model(&result).
		Where(`"bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ? `, linkId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve bank accounts by Id")
	}

	return result, nil
}

func (r *repositoryBase) GetBankAccount(bankAccountId uint64) (*models.BankAccount, error) {
	var result models.BankAccount
	err := r.txn.Model(&result).
		Where(`"bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."bank_account_id" = ? `, bankAccountId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve bank account")
	}

	return &result, nil
}

func (r *repositoryBase) UpdateBankAccounts(accounts []models.BankAccount) error {
	if len(accounts) == 0 {
		return nil
	}

	// Make sure each of the accounts has the correct accountId.
	for i := range accounts {
		accounts[i].AccountId = r.AccountId()
	}

	_, err := r.txn.Model(&accounts).WherePK().UpdateNotZero(&accounts)
	if err != nil {
		return errors.Wrap(err, "failed to update bank accounts")
	}

	return nil
}
