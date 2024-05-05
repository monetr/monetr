package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

var (
	ErrFundingScheduleNotFound = errors.New("funding schedule does not exist")
)

func (r *repositoryBase) GetFundingSchedules(ctx context.Context, bankAccountId ID[BankAccount]) ([]FundingSchedule, error) {
	span := sentry.StartSpan(ctx, "GetFundingSchedules")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
	}

	result := make([]FundingSchedule, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"funding_schedule"."account_id" = ?`, r.AccountId()).
		Where(`"funding_schedule"."bank_account_id" = ?`, bankAccountId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve funding schedules")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetFundingSchedule(ctx context.Context, bankAccountId ID[BankAccount], fundingScheduleId ID[FundingSchedule]) (*FundingSchedule, error) {
	span := sentry.StartSpan(ctx, "GetFundingSchedule")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":         r.AccountId(),
		"bankAccountId":     bankAccountId,
		"fundingScheduleId": fundingScheduleId,
	}

	var result FundingSchedule
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"funding_schedule"."account_id" = ?`, r.AccountId()).
		Where(`"funding_schedule"."bank_account_id" = ?`, bankAccountId).
		Where(`"funding_schedule"."funding_schedule_id" = ?`, fundingScheduleId).
		Limit(1).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "could not retrieve funding schedule")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}

func (r *repositoryBase) CreateFundingSchedule(ctx context.Context, fundingSchedule *FundingSchedule) error {
	span := sentry.StartSpan(ctx, "CreateFundingSchedule")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": fundingSchedule.BankAccountId,
	}

	fundingSchedule.AccountId = r.AccountId()

	if _, err := r.txn.ModelContext(span.Context(), fundingSchedule).Insert(fundingSchedule); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create funding schedule")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) UpdateFundingSchedule(ctx context.Context, fundingSchedule *FundingSchedule) error {
	span := sentry.StartSpan(ctx, "UpdateFundingSchedule")
	defer span.Finish()

	fundingSchedule.AccountId = r.AccountId()

	span.Data = map[string]interface{}{
		"accountId":         r.AccountId(),
		"fundingScheduleId": fundingSchedule.FundingScheduleId,
	}

	result, err := r.txn.ModelContext(span.Context(), fundingSchedule).
		WherePK().
		UpdateNotZero(&fundingSchedule)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update funding schedule")
	} else if result.RowsAffected() != 1 {
		span.Status = sentry.SpanStatusNotFound
		return errors.New("no rows updated")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) DeleteFundingSchedule(ctx context.Context, bankAccountId ID[BankAccount], fundingScheduleId ID[FundingSchedule]) error {
	span := sentry.StartSpan(ctx, "DeleteFundingSchedule")
	defer span.Finish()

	result, err := r.txn.ModelContext(span.Context(), &FundingSchedule{}).
		Where(`"funding_schedule"."account_id" = ?`, r.AccountId()).
		Where(`"funding_schedule"."bank_account_id" = ?`, bankAccountId).
		Where(`"funding_schedule"."funding_schedule_id" = ?`, fundingScheduleId).
		Delete()
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to remove funding schedule")
	} else if result.RowsAffected() == 0 {
		span.Status = sentry.SpanStatusNotFound
		return errors.WithStack(ErrFundingScheduleNotFound)
	}

	span.Status = sentry.SpanStatusOK
	return nil
}
