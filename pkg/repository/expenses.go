package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetExpenses(bankAccountId uint64) ([]models.Spending, error) {
	var result []models.Spending
	err := r.txn.Model(&result).
		Where(`"expense"."account_id" = ?`, r.AccountId()).
		Where(`"expense"."bank_account_id" = ?`, bankAccountId).
		Where(`"expense"."spending_type" = ?`, models.SpendingTypeExpense).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expenses")
	}

	return result, nil
}

func (r *repositoryBase) GetExpensesByFundingSchedule(bankAccountId, fundingScheduleId uint64) ([]models.Spending, error) {
	var result []models.Spending
	err := r.txn.Model(&result).
		Where(`"expense"."account_id" = ?`, r.AccountId()).
		Where(`"expense"."bank_account_id" = ?`, bankAccountId).
		Where(`"expense"."funding_schedule_id" = ?`, fundingScheduleId).
		Where(`"expense"."spending_type" = ?`, models.SpendingTypeExpense).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expenses for funding schedule")
	}

	return result, nil
}

func (r *repositoryBase) CreateExpense(expense *models.Spending) error {
	expense.AccountId = r.AccountId()
	expense.SpendingType = models.SpendingTypeExpense

	_, err := r.txn.Model(&expense).Insert(&expense)
	return errors.Wrap(err, "failed to create expense")
}

// UpdateExpenses should only be called with complete expense models. Do not use partial models with missing data for
// this action.
func (r *repositoryBase) UpdateExpenses(bankAccountId uint64, updates []models.Spending) error {
	for i := range updates {
		updates[i].AccountId = r.AccountId()
	}

	_, err := r.txn.Model(&updates).
		Where(`"expense"."account_id" = ?`, r.AccountId()).
		Where(`"expense"."bank_account_id" = ?`, bankAccountId).
		Where(`"expense"."spending_type" = ?`, models.SpendingTypeExpense).
		Update(&updates)
	if err != nil {
		return errors.Wrap(err, "failed to update expenses")
	}

	return nil
}

func (r *repositoryBase) GetExpense(bankAccountId, expenseId uint64) (*models.Spending, error) {
	var result models.Spending
	err := r.txn.Model(&result).
		Relation("FundingSchedule").
		Where(`"expense"."account_id" = ?`, r.AccountId()).
		Where(`"expense"."bank_account_id" = ?`, bankAccountId).
		Where(`"expense"."expense_id" = ?`, expenseId).
		Where(`"expense"."spending_type" = ?`, models.SpendingTypeExpense).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expense")
	}

	return &result, nil
}
