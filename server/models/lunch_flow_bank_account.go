package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
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
	bun.BaseModel `bun:"table:lunch_flow_bank_accounts,alias:lunch_flow_bank_account"`

	LunchFlowBankAccountId ID[LunchFlowBankAccount]           `json:"lunchFlowBankAccountId" bun:"lunch_flow_bank_account_id,notnull,pk"`
	AccountId              ID[Account]                        `json:"-" bun:"account_id,notnull,pk"`
	Account                *Account                           `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	LunchFlowLinkId        ID[LunchFlowLink]                  `json:"lunchFlowLinkId" bun:"lunch_flow_link_id,notnull,nullzero"`
	LunchFlowLink          *LunchFlowLink                     `json:"-" bun:"rel:belongs-to,join:lunch_flow_link_id=lunch_flow_link_id,join:account_id=account_id"`
	LunchFlowId            string                             `json:"lunchFlowId" bun:"lunch_flow_id,notnull,nullzero"`
	LunchFlowStatus        LunchFlowBankAccountExternalStatus `json:"lunchFlowStatus" bun:"lunch_flow_status,notnull,nullzero"`
	Name                   string                             `json:"name" bun:"name,notnull,nullzero"`
	InstitutionName        string                             `json:"institutionName" bun:"institution_name,notnull,nullzero"`
	Provider               string                             `json:"provider" bun:"provider,notnull,nullzero"`
	Currency               string                             `json:"currency" bun:"currency,notnull,nullzero"`
	Status                 LunchFlowBankAccountStatus         `json:"status" bun:"status,notnull,nullzero"`
	CurrentBalance         int64                              `json:"currentBalance" bun:"current_balance,notnull"`
	CreatedAt              time.Time                          `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy              ID[User]                           `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser          *User                              `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
	UpdatedAt              time.Time                          `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	DeletedAt              *time.Time                         `json:"deletedAt,omitempty" bun:"deleted_at"`
}

func (LunchFlowBankAccount) IdentityPrefix() string {
	return "lbac"
}

var (
	_ bun.BeforeAppendModelHook = (*LunchFlowBankAccount)(nil)
)

func (o *LunchFlowBankAccount) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.LunchFlowBankAccountId.IsZero() {
			o.LunchFlowBankAccountId = NewID[LunchFlowBankAccount]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}

		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = now
		}
	case *bun.UpdateQuery:
		o.UpdatedAt = time.Now()
	}

	return nil
}
