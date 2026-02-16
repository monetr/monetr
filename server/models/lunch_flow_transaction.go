package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type LunchFlowTransaction struct {
	tableName string `pg:"lunch_flow_transactions"`

	LunchFlowTransactionId ID[LunchFlowTransaction] `json:"lunchFlowTransactionId" pg:"lunch_flow_transaction_id,notnull,pk"`
	AccountId              ID[Account]              `json:"-" pg:"account_id,notnull,pk"`
	Account                *Account                 `json:"-" pg:"rel:has-one"`
	LunchFlowBankAccountId ID[LunchFlowBankAccount] `json:"-" pg:"lunch_flow_bank_account_id,notnull,unique:per_bank_account"`
	LunchFlowBankAccount   *LunchFlowBankAccount    `json:"-" pg:"rel:has-one"`
	LunchFlowId            string                   `json:"-" pg:"lunch_flow_id,notnull,unique:per_bank_account"`
	Merchant               string                   `json:"merchant" pg:"merchant"`
	Description            string                   `json:"description" pg:"description"`
	Date                   time.Time                `json:"date" pg:"date,notnull"`
	Currency               string                   `json:"currency" pg:"currency,notnull"`
	Amount                 int64                    `json:"amount" pg:"amount,notnull,use_zero"`
	IsPending              bool                     `json:"isPending" pg:"is_pending,notnull,use_zero"`
	CreatedAt              time.Time                `json:"createdAt" pg:"created_at,notnull,default:now()"`
	DeletedAt              *time.Time               `json:"deletedAt,omitempty" pg:"deleted_at"`
}

func (LunchFlowTransaction) IdentityPrefix() string {
	return "ltxn"
}

var (
	_ pg.BeforeInsertHook = (*PlaidTransaction)(nil)
)

func (o *LunchFlowTransaction) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.LunchFlowTransactionId.IsZero() {
		o.LunchFlowTransactionId = NewID[LunchFlowTransaction]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
