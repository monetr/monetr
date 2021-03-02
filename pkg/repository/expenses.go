package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetExpenses(bankAccountId uint64) ([]models.Expense, error) {
	var result []models.Expense
	err := r.txn.Model(&result).
		Where(`"expense"."account_id" = ?`, r.AccountId()).
		Where(`"expense"."bank_account_id" = ?`, bankAccountId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expenses")
	}

	return result, nil
}

func (r *repositoryBase) CreateExpense(expense *models.Expense) error {
	expense.AccountId = r.AccountId()

	_, err := r.txn.Model(&expense).Insert(&expense)
	return errors.Wrap(err, "failed to create expense")
}
