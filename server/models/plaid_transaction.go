package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type PlaidTransaction struct {
	tableName string `pg:"plaid_transactions"`

	PlaidTransactionId ID[PlaidTransaction] `json:"-" pg:"plaid_transaction_id,notnull,pk"`
	AccountId          ID[Account]          `json:"-" pg:"account_id,notnull,pk"`
	Account            *Account             `json:"-" pg:"rel:has-one"`
	PlaidBankAccountId ID[PlaidBankAccount] `json:"-" pg:"plaid_bank_account_id,notnull,unique:per_bank_account"`
	PlaidBankAccount   *PlaidBankAccount    `json:"-" pg:"rel:has-one"`
	PlaidId            string               `json:"-" pg:"plaid_id,notnull,unique:per_bank_account"`
	PendingPlaidId     *string              `json:"-" pg:"pending_plaid_id"`
	Categories         []string             `json:"categories" pg:"categories,type:'text[]'"`
	Category           *string              `json:"category" pg:"category"`
	Date               time.Time            `json:"date" pg:"date,notnull"`
	AuthorizedDate     *time.Time           `json:"authorizedDate" pg:"authorized_date"`
	Name               string               `json:"name,omitempty" pg:"name,notnull"`
	MerchantName       string               `json:"merchantName,omitempty" pg:"merchant_name"`
	Amount             int64                `json:"amount" pg:"amount,notnull,use_zero"`
	Currency           string               `json:"currency" pg:"currency,notnull"`
	IsPending          bool                 `json:"isPending" pg:"is_pending,notnull,use_zero"`
	CreatedAt          time.Time            `json:"createdAt" pg:"created_at,notnull,default:now()"`
	DeletedAt          *time.Time           `json:"deletedAt" pg:"deleted_at"`
}

func (PlaidTransaction) IdentityPrefix() string {
	return "ptxn"
}

var (
	_ pg.BeforeInsertHook = (*PlaidTransaction)(nil)
)

func (o *PlaidTransaction) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.PlaidTransactionId.IsZero() {
		o.PlaidTransactionId = NewID[PlaidTransaction]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
