package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type LunchFlowBankAccountStatus string

const (
	// LunchFlowBankAccountStatusActive means that the bank account can and will
	// automatically sync with Lunch Flow's API. This also means that the lunch
	// flow bank account has been associated with a monetr bank account record.
	LunchFlowBankAccountStatusActive LunchFlowBankAccountStatus = "active"
	// LunchFlowBankAccountStatusInactive means that the bank account does exist
	// in Lunch Flow's API, however it is not being actively synced with monetr.
	// Inactive items may not have a bank account in monetr associated with them.
	// Inactive items may or may not be associated with a monetr bank account
	// record. Depending on whether they were previously active and then
	// deactivated or whether they had never been activated.
	LunchFlowBankAccountStatusInactive LunchFlowBankAccountStatus = "inactive"
	// LunchFlowBankAccountStatusError means that sync attempts for this account
	// have failed and the account will no longer be automatically synced. For
	// data to continue to be synced the user must manually enable the account
	// again in the UI.
	LunchFlowBankAccountStatusError LunchFlowBankAccountStatus = "error"
)

type LunchFlowBankAccountExternalStatus string

const (
	LunchFlowBankAccountExternalStatusActive       LunchFlowBankAccountExternalStatus = "ACTIVE"
	LunchFlowBankAccountExternalStatusDisconnected LunchFlowBankAccountExternalStatus = "DISCONNECTED"
	LunchFlowBankAccountExternalStatusError        LunchFlowBankAccountExternalStatus = "ERROR"
)

type LunchFlowBankAccount struct {
	tableName string `pg:"lunch_flow_bank_accounts"`

	LunchFlowBankAccountId ID[LunchFlowBankAccount]           `json:"lunchFlowBankAccountId" pg:"lunch_flow_bank_account_id,notnull,pk"`
	AccountId              ID[Account]                        `json:"-" pg:"account_id,notnull,pk"`
	Account                *Account                           `json:"-" pg:"rel:has-one"`
	LunchFlowLinkId        ID[LunchFlowLink]                  `json:"lunchFlowLinkId" pg:"lunch_flow_link_id,notnull"`
	LunchFlowLink          *LunchFlowLink                     `json:"-" pg:"rel:has-one"`
	LunchFlowId            string                             `json:"lunchFlowId" pg:"lunch_flow_id,notnull"`
	LunchFlowStatus        LunchFlowBankAccountExternalStatus `json:"lunchFlowStatus" pg:"lunch_flow_status,notnull"`
	Name                   string                             `json:"name" pg:"name,notnull"`
	InstitutionName        string                             `json:"institutionName" pg:"institution_name,notnull"`
	Provider               string                             `json:"provider" pg:"provider,notnull"`
	Currency               string                             `json:"currency" pg:"currency,notnull"`
	Status                 LunchFlowBankAccountStatus         `json:"status" pg:"status,notnull"`
	CurrentBalance         int64                              `json:"currentBalance" pg:"current_balance,notnull,use_zero"`
	CreatedAt              time.Time                          `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy              ID[User]                           `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser          *User                              `json:"-" pg:"rel:has-one,fk:created_by"`
	UpdatedAt              time.Time                          `json:"updatedAt" pg:"updated_at,notnull"`
	DeletedAt              *time.Time                         `json:"deletedAt,omitempty" pg:"deleted_at"`
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
