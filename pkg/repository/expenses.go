package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
	"time"
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

func (r *repositoryBase) GetExpensesByFundingSchedule(bankAccountId, fundingScheduleId uint64) ([]models.Expense, error) {
	var result []models.Expense
	err := r.txn.Model(&result).
		Where(`"expense"."account_id" = ?`, r.AccountId()).
		Where(`"expense"."bank_account_id" = ?`, bankAccountId).
		Where(`"expense"."funding_schedule_id" = ?`, fundingScheduleId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expenses for funding schedule")
	}

	return result, nil
}

func (r *repositoryBase) CreateExpense(expense *models.Expense) error {
	expense.AccountId = r.AccountId()

	_, err := r.txn.Model(&expense).Insert(&expense)
	return errors.Wrap(err, "failed to create expense")
}

type ExpenseUpdateItem struct {
	ExpenseId              uint64     `pg:"expense_id"`
	CurrentAmount          int64      `pg:"current_amount"`
	NextContributionAmount int64      `pg:"next_contribution_amount"`
	IsBehind               bool       `pg:"is_behind"`
	LastRecurrence         *time.Time `pg:"last_recurrence"`
	NextRecurrence         *time.Time `pg:"next_recurrence"`
}

func (r *repositoryBase) UpdateExpenseBalances(bankAccountId uint64, updates []ExpenseUpdateItem) error {
	return nil
}
