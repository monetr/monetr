package repository

import "github.com/pkg/errors"

type Balances struct {
	tableName string `pg:"balances"`

	BankAccountId uint64 `json:"bankAccountId" pg:"bank_account_id"`
	AccountId     uint64 `json:"-" pg:"account_id"`
	Current       int64  `json:"current" pg:"current"`
	Available     int64  `json:"available" pg:"available"`
	Safe          int64  `json:"safe" pg:"safe"`
	Expenses      int64  `json:"expenses" pg:"expenses"`
	Goals         int64  `json:"goals" pg:"goals"`
}

func (r *repositoryBase) GetBalances(bankAccountId uint64) (*Balances, error) {
	var balance Balances
	err := r.txn.Model(&balance).
		Where(`"balances"."account_id" = ?`, r.AccountId()).
		Where(`"balances"."bank_account_id" = ?`, bankAccountId).
		Limit(1).
		Select(&balance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve balances")
	}

	return &balance, nil
}
