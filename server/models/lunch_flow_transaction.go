package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type LunchFlowTransaction struct {
	bun.BaseModel `bun:"table:lunch_flow_transactions,alias:lunch_flow_transaction"`

	LunchFlowTransactionId ID[LunchFlowTransaction] `json:"lunchFlowTransactionId" bun:"lunch_flow_transaction_id,notnull,pk"`
	AccountId              ID[Account]              `json:"-" bun:"account_id,notnull,pk"`
	Account                *Account                 `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	LunchFlowBankAccountId ID[LunchFlowBankAccount] `json:"-" bun:"lunch_flow_bank_account_id,notnull,unique:per_bank_account,nullzero"`
	LunchFlowBankAccount   *LunchFlowBankAccount    `json:"-" bun:"rel:belongs-to,join:lunch_flow_bank_account_id=lunch_flow_bank_account_id,join:account_id=account_id"`
	LunchFlowId            string                   `json:"-" bun:"lunch_flow_id,notnull,unique:per_bank_account,nullzero"`
	Merchant               string                   `json:"merchant" bun:"merchant,nullzero"`
	Description            string                   `json:"description" bun:"description,nullzero"`
	Date                   time.Time                `json:"date" bun:"date,notnull,nullzero"`
	Currency               string                   `json:"currency" bun:"currency,notnull,nullzero"`
	Amount                 int64                    `json:"amount" bun:"amount,notnull"`
	IsPending              bool                     `json:"isPending" bun:"is_pending,notnull"`
	CreatedAt              time.Time                `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
	DeletedAt              *time.Time               `json:"deletedAt,omitempty" bun:"deleted_at"`
}

func (LunchFlowTransaction) IdentityPrefix() string {
	return "ltxn"
}

var (
	_ bun.BeforeAppendModelHook = (*LunchFlowTransaction)(nil)
)

func (o *LunchFlowTransaction) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.LunchFlowTransactionId.IsZero() {
			o.LunchFlowTransactionId = NewID[LunchFlowTransaction]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}
