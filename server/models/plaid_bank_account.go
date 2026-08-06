package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type PlaidBankAccount struct {
	bun.BaseModel `bun:"table:plaid_bank_accounts,alias:plaid_bank_account"`

	PlaidBankAccountId ID[PlaidBankAccount] `json:"-" bun:"plaid_bank_account_id,notnull,pk"`
	AccountId          ID[Account]          `json:"-" bun:"account_id,notnull,pk"`
	Account            *Account             `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	PlaidLinkId        ID[PlaidLink]        `json:"-" bun:"plaid_link_id,nullzero"`
	PlaidLink          *PlaidLink           `json:"-" bun:"rel:belongs-to,join:plaid_link_id=plaid_link_id"`
	PlaidId            string               `json:"-" bun:"plaid_id,notnull,nullzero"`
	Name               string               `json:"name" bun:"name,notnull,nullzero"`
	OfficialName       string               `json:"officialName" bun:"official_name,nullzero"`
	Mask               string               `json:"mask" bun:"mask,nullzero"`
	Currency           string               `json:"currency" bun:"currency,notnull,nullzero"`
	AvailableBalance   int64                `json:"availableBalance" bun:"available_balance,notnull"`
	CurrentBalance     int64                `json:"currentBalance" bun:"current_balance,notnull"`
	LimitBalance       int64                `json:"limitBalance" bun:"limit_balance"`
	CreatedAt          time.Time            `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy          ID[User]             `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser      *User                `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
}

func (PlaidBankAccount) IdentityPrefix() string {
	return "pbac"
}

var (
	_ bun.BeforeAppendModelHook = (*PlaidBankAccount)(nil)
)

func (o *PlaidBankAccount) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.PlaidBankAccountId.IsZero() {
			o.PlaidBankAccountId = NewID[PlaidBankAccount]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}
