package repository

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type Balances struct {
	BankAccountId ID[BankAccount] `json:"bankAccountId" bun:"bank_account_id"`
	AccountId     ID[Account]     `json:"-" bun:"account_id"`
	Currency      string          `json:"currency" bun:"currency"`
	Current       int64           `json:"current" bun:"current"`
	Available     int64           `json:"available" bun:"available"`
	Limit         int64           `json:"limit" bun:"limit"`
	Free          int64           `json:"free" bun:"free"`
	Expenses      int64           `json:"expenses" bun:"expenses"`
	Goals         int64           `json:"goals" bun:"goals"`
}

func (r *repositoryBase) GetBalances(
	ctx context.Context,
	bankAccountId ID[BankAccount],
) (*Balances, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	var balance Balances

	expenseQuery := r.txn.NewSelect().Model(&Spending{}).
		ColumnExpr(`"spending"."bank_account_id"`).
		ColumnExpr(`"spending"."account_id"`).
		ColumnExpr(`SUM("spending"."current_amount") AS "current_amount"`).
		Where(`"spending"."spending_type" = ?`, SpendingTypeExpense).
		GroupExpr(`"spending"."bank_account_id"`).
		GroupExpr(`"spending"."account_id"`)

	goalQuery := r.txn.NewSelect().Model(&Spending{}).
		ColumnExpr(`"spending"."bank_account_id"`).
		ColumnExpr(`"spending"."account_id"`).
		ColumnExpr(`SUM("spending"."current_amount") AS "current_amount"`).
		Where(`"spending"."spending_type" = ?`, SpendingTypeGoal).
		GroupExpr(`"spending"."bank_account_id"`).
		GroupExpr(`"spending"."account_id"`)

	err := r.txn.NewSelect().Model(&BankAccount{}).
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
		Scan(span.Context(), &balance)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve balances")
	}

	span.Status = sentry.SpanStatusOK

	return &balance, nil
}
