package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type LunchFlowBankAccountStatus string

const (
	LunchFlowBankAccountStatusActive       LunchFlowBankAccountStatus = "active"
	LunchFlowBankAccountStatusDisconnected LunchFlowBankAccountStatus = "disconnected"
	LunchFlowBankAccountStatusError        LunchFlowBankAccountStatus = "error"
)

type LunchFlowBankAccount struct {
	tableName string `pg:"lunchflow_bank_accounts"`

	LunchFlowBankAccountId ID[LunchFlowBankAccount]   `json:"lunchFlowBankAccountId" pg:"lunchflow_bank_account_id,notnull,pk"`
	AccountId              ID[Account]                `json:"-" pg:"account_id,notnull,pk"`
	Account                *Account                   `json:"-" pg:"rel:has-one"`
	LunchFlowLinkId        ID[LunchFlowLink]          `json:"lunchFlowLinkId" pg:"lunchflow_link_id,notnull"`
	LunchFlowLink          *LunchFlowLink             `json:"-" pg:"rel:has-one"`
	LunchFlowId            string                     `json:"lunchFlowId" pg:"lunchflow_id,notnull"`
	Name                   string                     `json:"name" pg:"name,notnull"`
	InstitutionName        string                     `json:"institutionName" pg:"institution_name,notnull"`
	Provider               string                     `json:"provider" pg:"provider,notnull"`
	Currency               string                     `json:"currency" pg:"currency,notnull"`
	Status                 LunchFlowBankAccountStatus `json:"status" pg:"status,notnull"`
	CurrentBalance         int64                      `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	CreatedAt              time.Time                  `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy              ID[User]                   `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser          *User                      `json:"-" pg:"rel:has-one,fk:created_by"`
	UpdatedAt              time.Time                  `json:"updatedAt" pg:"updated_at,notnull"`
	DeletedAt              *time.Time                 `json:"deletedAt" pg:"deleted_at"`
}

func (LunchFlowBankAccount) IdentityPrefix() string {
	return "lbac"
}

var (
	_ pg.BeforeInsertHook = (*LunchFlowBankAccount)(nil)
	_ pg.BeforeUpdateHook = (*LunchFlowBankAccount)(nil)
)

func (o *LunchFlowBankAccount) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.LunchFlowBankAccountId.IsZero() {
		o.LunchFlowBankAccountId = NewID(o)
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = now
	}

	return ctx, nil
}

func (o *LunchFlowBankAccount) BeforeUpdate(ctx context.Context) (context.Context, error) {
	o.UpdatedAt = time.Now()
	return ctx, nil
}
