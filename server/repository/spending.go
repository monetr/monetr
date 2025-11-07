package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

func (r *repositoryBase) GetSpending(ctx context.Context, bankAccountId ID[BankAccount]) ([]Spending, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
	}

	result := make([]Spending, 0)
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

func (r *repositoryBase) GetSpendingExists(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending]) (bool, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingId":    spendingId,
	}

	ok, err := r.txn.ModelContext(span.Context(), &Spending{}).
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

func (r *repositoryBase) GetSpendingByFundingSchedule(ctx context.Context, bankAccountId ID[BankAccount], fundingScheduleId ID[FundingSchedule]) ([]Spending, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":         r.AccountId(),
		"bankAccountId":     bankAccountId,
		"fundingScheduleId": fundingScheduleId,
	}

	result := make([]Spending, 0)
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

func (r *repositoryBase) CreateSpending(ctx context.Context, spending *Spending) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": spending.BankAccountId,
	}

	spending.AccountId = r.AccountId()
	spending.CreatedAt = r.clock.Now().UTC()

	if _, err := r.txn.ModelContext(span.Context(), spending).Insert(spending); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create spending")
	}

	span.Data["spendingId"] = spending.SpendingId
	span.Status = sentry.SpanStatusOK

	return nil
}

// UpdateSpending should only be called with complete expense  Do not use partial models with missing data for
// this action.
func (r *repositoryBase) UpdateSpending(ctx context.Context, bankAccountId ID[BankAccount], updates []Spending) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	spendingIds := make([]ID[Spending], len(updates))
	for i := range updates {
		updates[i].AccountId = r.AccountId()
		updates[i].BankAccountId = bankAccountId
		spendingIds[i] = updates[i].SpendingId
	}

	span.Data = map[string]any{
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

func (r *repositoryBase) GetSpendingById(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending]) (*Spending, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingId":    spendingId,
	}

	var result Spending
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

func (r *repositoryBase) DeleteSpending(ctx context.Context, bankAccountId ID[BankAccount], spendingId ID[Spending]) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingId":    spendingId,
	}

	_, err := r.txn.ModelContext(span.Context(), &Transaction{}).
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

	result, err := r.txn.ModelContext(span.Context(), &Spending{}).
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
