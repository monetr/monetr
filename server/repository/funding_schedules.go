package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

var (
	ErrFundingScheduleNotFound = errors.New("funding schedule does not exist")
)

func (r *repositoryBase) GetFundingSchedules(
	ctx context.Context,
	bankAccountId ID[BankAccount],
) ([]FundingSchedule, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)

	result := make([]FundingSchedule, 0)
	err := r.txn.NewSelect().Model(&result).
		Where(`"funding_schedule"."account_id" = ?`, r.AccountId()).
		Where(`"funding_schedule"."bank_account_id" = ?`, bankAccountId).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve funding schedules")
	}

	span.Status = sentry.SpanStatusOK

	return result, nil
}

func (r *repositoryBase) GetFundingSchedule(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	fundingScheduleId ID[FundingSchedule],
) (*FundingSchedule, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", bankAccountId)
	span.SetData("fundingScheduleId", fundingScheduleId)

	var result FundingSchedule
	err := r.txn.NewSelect().Model(&result).
		Where(`"funding_schedule"."account_id" = ?`, r.AccountId()).
		Where(`"funding_schedule"."bank_account_id" = ?`, bankAccountId).
		Where(`"funding_schedule"."funding_schedule_id" = ?`, fundingScheduleId).
		Limit(1).
		Scan(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "could not retrieve funding schedule")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}

func (r *repositoryBase) CreateFundingSchedule(
	ctx context.Context,
	fundingSchedule *FundingSchedule,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", fundingSchedule.BankAccountId)

	fundingSchedule.AccountId = r.AccountId()

	if _, err := r.txn.NewInsert().
		Model(fundingSchedule).
		Returning("*").
		Exec(span.Context()); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create funding schedule")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) UpdateFundingSchedule(ctx context.Context, fundingSchedule *FundingSchedule) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()
	span.SetData("accountId", r.AccountId())
	span.SetData("bankAccountId", fundingSchedule.BankAccountId)
	span.SetData("fundingScheduleId", fundingSchedule.FundingScheduleId)

	fundingSchedule.AccountId = r.AccountId()

	result, err := r.txn.NewUpdate().Model(fundingSchedule).
		WherePK().
		// go-pg's UpdateNotZero skipped zero-valued fields EXCEPT those tagged
		// use_zero. bun's OmitZero has no such override, so the formerly-use_zero
		// boolean columns are forced through with explicit values; without this a
		// PATCH turning e.g. excludeWeekends off would silently not persist.
		OmitZero().
		Value("exclude_weekends", "?", fundingSchedule.ExcludeWeekends).
		Value("wait_for_deposit", "?", fundingSchedule.WaitForDeposit).
		Value("auto_create_transaction", "?", fundingSchedule.AutoCreateTransaction).
		Returning("*").
		Exec(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update funding schedule")
	} else if affected, err := result.RowsAffected(); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update funding schedule")
	} else if affected != 1 {
		span.Status = sentry.SpanStatusNotFound
		return errors.New("no rows updated")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) DeleteFundingSchedule(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	fundingScheduleId ID[FundingSchedule],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result, err := r.txn.NewDelete().Model(&FundingSchedule{}).
		Where(`"funding_schedule"."account_id" = ?`, r.AccountId()).
		Where(`"funding_schedule"."bank_account_id" = ?`, bankAccountId).
		Where(`"funding_schedule"."funding_schedule_id" = ?`, fundingScheduleId).
		Exec(span.Context())
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to remove funding schedule")
	} else if affected, err := result.RowsAffected(); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to remove funding schedule")
	} else if affected == 0 {
		span.Status = sentry.SpanStatusNotFound
		return errors.WithStack(ErrFundingScheduleNotFound)
	}

	span.Status = sentry.SpanStatusOK
	return nil
}
