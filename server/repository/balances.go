package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type Balances struct {
	tableName string `pg:"balances"`

	BankAccountId ID[BankAccount] `json:"bankAccountId" pg:"bank_account_id"`
	AccountId     ID[Account]     `json:"-" pg:"account_id"`
	Current       int64           `json:"current" pg:"current"`
	Available     int64           `json:"available" pg:"available"`
	Limit         int64           `json:"limit" pg:"limit"`
	Free          int64           `json:"free" pg:"free"`
	Expenses      int64           `json:"expenses" pg:"expenses"`
	Goals         int64           `json:"goals" pg:"goals"`
}

func (r *repositoryBase) GetBalances(ctx context.Context, bankAccountId ID[BankAccount]) (*Balances, error) {
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
