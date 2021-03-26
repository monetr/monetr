package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetExpenses(bankAccountId uint64) ([]models.Spending, error) {
	var result []models.Spending
	err := r.txn.Model(&result).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."spending_type" = ?`, models.SpendingTypeExpense).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expenses")
	}

	return result, nil
}

func (r *repositoryBase) GetExpensesByFundingSchedule(bankAccountId, fundingScheduleId uint64) ([]models.Spending, error) {
	var result []models.Spending
	err := r.txn.Model(&result).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."funding_schedule_id" = ?`, fundingScheduleId).
		Where(`"spending"."spending_type" = ?`, models.SpendingTypeExpense).
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
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."spending_type" = ?`, models.SpendingTypeExpense).
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
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."expense_id" = ?`, expenseId).
		Where(`"spending"."spending_type" = ?`, models.SpendingTypeExpense).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expense")
	}

	return &result, nil
}
