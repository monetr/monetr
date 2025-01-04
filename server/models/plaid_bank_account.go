package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type PlaidBankAccount struct {
	tableName string `pg:"plaid_bank_accounts"`

	PlaidBankAccountId ID[PlaidBankAccount] `json:"-" pg:"plaid_bank_account_id,notnull,pk"`
	AccountId          ID[Account]          `json:"-" pg:"account_id,notnull,pk"`
	Account            *Account             `json:"-" pg:"rel:has-one"`
	PlaidLinkId        ID[PlaidLink]        `json:"-" pg:"plaid_link_id"`
	PlaidLink          *PlaidLink           `json:"-" pg:"rel:has-one"`
	PlaidId            string               `json:"-" pg:"plaid_id,notnull"`
	Name               string               `json:"name" pg:"name,notnull"`
	OfficialName       string               `json:"officialName" pg:"official_name"`
	Mask               string               `json:"mask" pg:"mask"`
	AvailableBalance   int64                `json:"availableBalance" pg:"available_balance,notnull,use_zero"`
	CurrentBalance     int64                `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	LimitBalance       int64                `json:"limitBalance" pg:"limit_balance,use_zero"`
	CreatedAt          time.Time            `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy          ID[User]             `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser      *User                `json:"-" pg:"rel:has-one,fk:created_by"`
}

func (PlaidBankAccount) IdentityPrefix() string {
	return "pbac"
}

var (
	_ pg.BeforeInsertHook = (*PlaidBankAccount)(nil)
)

func (o *PlaidBankAccount) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.PlaidBankAccountId.IsZero() {
		o.PlaidBankAccountId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}
