package repository

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetSpending(ctx context.Context, bankAccountId uint64) ([]models.Spending, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
	}

	result := make([]models.Spending, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve spending")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetSpendingExists(ctx context.Context, bankAccountId, spendingId uint64) (bool, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingId":    spendingId,
	}

	ok, err := r.txn.ModelContext(span.Context(), &models.Spending{}).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."spending_id" = ?`, spendingId).
		Limit(1).
		Exists()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
	} else {
		span.Status = sentry.SpanStatusOK
	}

	return ok, errors.Wrap(err, "failed to verify spending object exists")
}

func (r *repositoryBase) GetSpendingByFundingSchedule(ctx context.Context, bankAccountId, fundingScheduleId uint64) ([]models.Spending, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":         r.AccountId(),
		"bankAccountId":     bankAccountId,
		"fundingScheduleId": fundingScheduleId,
	}

	result := make([]models.Spending, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."funding_schedule_id" = ?`, fundingScheduleId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve expenses for funding schedule")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) CreateSpending(ctx context.Context, spending *models.Spending) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": spending.BankAccountId,
	}

	spending.AccountId = r.AccountId()
	spending.DateCreated = time.Now().UTC()

	if _, err := r.txn.ModelContext(span.Context(), spending).Insert(spending); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create spending")
	}

	span.Data["spendingId"] = spending.SpendingId
	span.Status = sentry.SpanStatusOK

	return nil
}

// UpdateSpending should only be called with complete expense models. Do not use partial models with missing data for
// this action.
func (r *repositoryBase) UpdateSpending(ctx context.Context, bankAccountId uint64, updates []models.Spending) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	spendingIds := make([]uint64, len(updates))
	for i := range updates {
		updates[i].AccountId = r.AccountId()
		updates[i].BankAccountId = bankAccountId
		spendingIds[i] = updates[i].SpendingId
	}

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingIds":   spendingIds,
	}

	_, err := r.txn.ModelContext(span.Context(), &updates).
		Update(&updates)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update expenses")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) GetSpendingById(ctx context.Context, bankAccountId, spendingId uint64) (*models.Spending, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingId":    spendingId,
	}

	var result models.Spending
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("FundingSchedule").
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."spending_id" = ?`, spendingId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve expense")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}

func (r *repositoryBase) DeleteSpending(ctx context.Context, bankAccountId, spendingId uint64) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingId":    spendingId,
	}

	_, err := r.txn.ModelContext(span.Context(), &models.Transaction{}).
		Set(`"spending_id" = NULL`).
		Set(`"spending_amount" = NULL`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."spending_id" = ?`, spendingId).
		Update()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to remove spending from any transactions")
	}

	result, err := r.txn.ModelContext(span.Context(), &models.Spending{}).
		Where(`"spending"."account_id" = ?`, r.AccountId()).
		Where(`"spending"."bank_account_id" = ?`, bankAccountId).
		Where(`"spending"."spending_id" = ?`, spendingId).
		Delete()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to delete spending")
	}

	span.Status = sentry.SpanStatusOK

	if result.RowsAffected() != 1 {
		span.Status = sentry.SpanStatusDataLoss
		return errors.Errorf("invalid number of spending(s) deleted: %d", result.RowsAffected())
	}

	span.Status = sentry.SpanStatusOK

	return nil
}
