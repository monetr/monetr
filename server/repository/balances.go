package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/pkg/errors"
)

type Balances struct {
	tableName string `pg:"balances"`

	BankAccountId uint64 `json:"bankAccountId" pg:"bank_account_id"`
	AccountId     uint64 `json:"-" pg:"account_id"`
	Current       int64  `json:"current" pg:"current"`
	Available     int64  `json:"available" pg:"available"`
	Free          int64  `json:"free" pg:"free"`
	Expenses      int64  `json:"expenses" pg:"expenses"`
	Goals         int64  `json:"goals" pg:"goals"`
}

func (r *repositoryBase) GetBalances(ctx context.Context, bankAccountId uint64) (*Balances, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var balance Balances
	err := r.txn.ModelContext(span.Context(), &balance).
		Where(`"balances"."account_id" = ?`, r.AccountId()).
		Where(`"balances"."bank_account_id" = ?`, bankAccountId).
		Limit(1).
		Select(&balance)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve balances")
	}

	span.Status = sentry.SpanStatusOK

	return &balance, nil
}

type FundingStats struct {
	tableName string `pg:"funding_stats"`

	AccountId               uint64 `json:"-" pg:"account_id"`
	BankAccountId           uint64 `json:"bankAccountId" pg:"bank_account_id"`
	FundingScheduleId       uint64 `json:"fundingScheduleId" pg:"funding_schedule_id"`
	NextExpenseContribution int64  `json:"nextExpenseContribution" pg:"next_expense_contribution"`
	NextGoalContribution    int64  `json:"nextGoalContribution" pg:"next_goal_contribution"`
}

func (r *repositoryBase) GetFundingStats(ctx context.Context, bankAccountId uint64) ([]FundingStats, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	stats := make([]FundingStats, 0)
	err := r.txn.ModelContext(span.Context(), &stats).
		Where(`"funding_stats"."account_id" = ?`, r.AccountId()).
		Where(`"funding_stats"."bank_account_id" = ?`, bankAccountId).
		Select(&stats)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve funding status")
	}

	return stats, nil
}
