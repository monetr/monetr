package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type PlaidTransaction struct {
	bun.BaseModel `bun:"table:plaid_transactions,alias:plaid_transaction"`

	PlaidTransactionId ID[PlaidTransaction] `json:"-" bun:"plaid_transaction_id,notnull,pk"`
	AccountId          ID[Account]          `json:"-" bun:"account_id,notnull,pk"`
	Account            *Account             `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	PlaidBankAccountId ID[PlaidBankAccount] `json:"-" bun:"plaid_bank_account_id,notnull,unique:per_bank_account,nullzero"`
	PlaidBankAccount   *PlaidBankAccount    `json:"-" bun:"rel:belongs-to,join:plaid_bank_account_id=plaid_bank_account_id,join:account_id=account_id"`
	PlaidId            string               `json:"-" bun:"plaid_id,notnull,unique:per_bank_account,nullzero"`
	PendingPlaidId     *string              `json:"-" bun:"pending_plaid_id"`
	Categories         []string             `json:"categories" bun:"categories,array,nullzero"`
	Category           *string              `json:"category" bun:"category"`
	Date               time.Time            `json:"date" bun:"date,notnull,nullzero"`
	AuthorizedDate     *time.Time           `json:"authorizedDate" bun:"authorized_date"`
	Name               string               `json:"name,omitempty" bun:"name,notnull,nullzero"`
	MerchantName       string               `json:"merchantName,omitempty" bun:"merchant_name,nullzero"`
	Amount             int64                `json:"amount" bun:"amount,notnull"`
	Currency           string               `json:"currency" bun:"currency,notnull,nullzero"`
	IsPending          bool                 `json:"isPending" bun:"is_pending,notnull"`
	CreatedAt          time.Time            `json:"createdAt" bun:"created_at,notnull,default:now(),nullzero"`
	DeletedAt          *time.Time           `json:"deletedAt" bun:"deleted_at"`
}

func (PlaidTransaction) IdentityPrefix() string {
	return "ptxn"
}

var (
	_ bun.BeforeAppendModelHook = (*PlaidTransaction)(nil)
)

func (o *PlaidTransaction) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.PlaidTransactionId.IsZero() {
			o.PlaidTransactionId = NewID[PlaidTransaction]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}
