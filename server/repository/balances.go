package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type Balances struct {
	BankAccountId ID[BankAccount] `json:"bankAccountId" pg:"bank_account_id"`
	AccountId     ID[Account]     `json:"-" pg:"account_id"`
	Currency      string          `json:"currency" pg:"currency"`
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

	expenseQuery := r.txn.ModelContext(span.Context(), &Spending{}).
		ColumnExpr(`"spending"."bank_account_id"`).
		ColumnExpr(`"spending"."account_id"`).
		ColumnExpr(`SUM("spending"."current_amount") AS "current_amount"`).
		Where(`"spending"."spending_type" = ?`, SpendingTypeExpense).
		GroupExpr(`"spending"."bank_account_id"`).
		GroupExpr(`"spending"."account_id"`)

	goalQuery := r.txn.ModelContext(span.Context(), &Spending{}).
		ColumnExpr(`"spending"."bank_account_id"`).
		ColumnExpr(`"spending"."account_id"`).
		ColumnExpr(`SUM("spending"."current_amount") AS "current_amount"`).
		Where(`"spending"."spending_type" = ?`, SpendingTypeGoal).
		GroupExpr(`"spending"."bank_account_id"`).
		GroupExpr(`"spending"."account_id"`)

	err := r.txn.ModelContext(span.Context(), &BankAccount{}).
		With(`expense`, expenseQuery).
		With(`goal`, goalQuery).
		ColumnExpr(`"bank_account"."bank_account_id"`).
		ColumnExpr(`"bank_account"."account_id"`).
		ColumnExpr(`"bank_account"."currency"`).
		ColumnExpr(`"bank_account"."current_balance" AS "current"`).
		ColumnExpr(`"bank_account"."available_balance" AS "available"`).
		ColumnExpr(`"bank_account"."limit_balance" AS "limit"`).
		ColumnExpr(`"bank_account"."available_balance" - SUM(COALESCE("expense"."current_amount", 0)) - SUM(COALESCE("goal"."current_amount", 0)) AS "free"`).
		ColumnExpr(`SUM(COALESCE("expense"."current_amount", 0)) AS "expenses"`).
		ColumnExpr(`SUM(COALESCE("goal"."current_amount", 0)) AS "goals"`).
		Join(`LEFT JOIN "expense"`).
		JoinOn(`"expense"."bank_account_id" = "bank_account"."bank_account_id"`).
		JoinOn(`"expense"."account_id" = "bank_account"."account_id"`).
		Join(`LEFT JOIN "goal"`).
		JoinOn(`"goal"."bank_account_id" = "bank_account"."bank_account_id"`).
		JoinOn(`"goal"."account_id" = "bank_account"."account_id"`).
		Where(`"bank_account"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."bank_account_id" = ?`, bankAccountId).
		GroupExpr(`"bank_account"."bank_account_id"`).
		GroupExpr(`"bank_account"."account_id"`).
		Limit(1).
		Select(&balance)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve balances")
	}

	span.Status = sentry.SpanStatusOK

	return &balance, nil
}
