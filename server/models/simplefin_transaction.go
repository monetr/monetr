package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type SimpleFinTransaction struct {
	bun.BaseModel `bun:"table:simplefin_transactions,alias:simplefin_transaction"`

	SimpleFinTransactionId ID[SimpleFinTransaction] `json:"simpleFinTransactionId" bun:"simplefin_transaction_id,notnull,pk"`
	AccountId              ID[Account]              `json:"-" bun:"account_id,notnull,pk"`
	Account                *Account                 `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	SimpleFinBankAccountId ID[SimpleFinBankAccount] `json:"-" bun:"simplefin_bank_account_id,notnull,unique:per_bank_account,nullzero"`
	SimpleFinBankAccount   *SimpleFinBankAccount    `json:"-" bun:"rel:belongs-to,join:simplefin_bank_account_id=simplefin_bank_account_id,join:account_id=account_id"`
	SimpleFinId            string                   `json:"-" bun:"simplefin_id,notnull,unique:per_bank_account,nullzero"`
	// Description is not nullzero because SimpleFIN might give us an empty one
	// and the column is NOT NULL.
	Description  string     `json:"description" bun:"description,notnull"`
	Amount       int64      `json:"amount" bun:"amount,notnull"`
	Posted       time.Time  `json:"posted" bun:"posted,notnull,nullzero"`
	TransactedAt *time.Time `json:"transactedAt" bun:"transacted_at"`
	CreatedAt    time.Time  `json:"createdAt" bun:"created_at,notnull,nullzero"`
	DeletedAt    *time.Time `json:"deletedAt,omitempty" bun:"deleted_at"`
}

func (SimpleFinTransaction) IdentityPrefix() string {
	return "stxn"
}

var (
	_ bun.BeforeAppendModelHook = (*SimpleFinTransaction)(nil)
)

func (o *SimpleFinTransaction) BeforeAppendModel(
	ctx context.Context,
	query bun.Query,
) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.SimpleFinTransactionId.IsZero() {
			o.SimpleFinTransactionId = NewID[SimpleFinTransaction]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}
