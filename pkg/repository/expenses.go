package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/pkg/errors"
	"time"
)

func (r *repositoryBase) GetSpending(ctx context.Context, bankAccountId uint64) ([]models.Spending, error) {
	var result []models.Spending
	err := r.txn.ModelContext(ctx, &result).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve spending")
	}

	return result, nil
}

func (r *repositoryBase) GetSpendingByFundingSchedule(bankAccountId, fundingScheduleId uint64) ([]models.Spending, error) {
	result := make([]models.Spending, 0)
	err := r.txn.Model(&result).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."funding_schedule_id" = ?`, fundingScheduleId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expenses for funding schedule")
	}

	return result, nil
}

func (r *repositoryBase) CreateSpending(spending *models.Spending) error {
	spending.AccountId = r.AccountId()
	spending.DateCreated = time.Now().UTC()

	_, err := r.txn.Model(spending).Insert(spending)
	return errors.Wrap(err, "failed to create spending")
}

// UpdateExpenses should only be called with complete expense models. Do not use partial models with missing data for
// this action.
func (r *repositoryBase) UpdateExpenses(bankAccountId uint64, updates []models.Spending) error {
	for i := range updates {
		updates[i].AccountId = r.AccountId()
		updates[i].BankAccountId = bankAccountId
	}

	_, err := r.txn.Model(&updates).
		Update(&updates)
	if err != nil {
		return errors.Wrap(err, "failed to update expenses")
	}

	return nil
}

func (r *repositoryBase) GetSpendingById(bankAccountId, spendingId uint64) (*models.Spending, error) {
	var result models.Spending
	err := r.txn.Model(&result).
		Relation("FundingSchedule").
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."spending_id" = ?`, spendingId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve expense")
	}

	return &result, nil
}

func (r *repositoryBase) DeleteSpending(ctx context.Context, bankAccountId, spendingId uint64) error {
	span := sentry.StartSpan(ctx, "Delete Spending")
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &models.Transaction{}).
		Set(`"spending_id" = NULL`).
		Set(`"spending_amount" = NULL`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."spending_id" = ?`, spendingId).
		Update()
	if err != nil {
		return errors.Wrap(err, "failed to remove spending from any transactions")
	}

	result, err := r.txn.ModelContext(span.Context(), &models.Spending{}).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."spending_id" = ?`, spendingId).
		Delete()
	if err != nil {
		return errors.Wrap(err, "failed to delete spending")
	}

	if result.RowsAffected() != 1 {
		return errors.Errorf("invalid number of spending(s) deleted: %d", result.RowsAffected())
	}

	return nil
}
