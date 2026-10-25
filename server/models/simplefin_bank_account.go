package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type SimpleFinBankAccountStatus string

const (
	// SimpleFinBankAccountStatusActive means that the bank account will be synced
	// from SimpleFIN and has a monetr bank account associated with it.
	SimpleFinBankAccountStatusActive SimpleFinBankAccountStatus = "active"
	// SimpleFinBankAccountStatusInactive means the account exists in SimpleFIN
	// but is not being synced. It may or may not have a monetr bank account
	// associated with it, depending on whether it was active before.
	SimpleFinBankAccountStatusInactive SimpleFinBankAccountStatus = "inactive"
)

type SimpleFinBankAccount struct {
	bun.BaseModel `bun:"table:simplefin_bank_accounts,alias:simplefin_bank_account"`

	SimpleFinBankAccountId ID[SimpleFinBankAccount]   `json:"simpleFinBankAccountId" bun:"simplefin_bank_account_id,notnull,pk"`
	AccountId              ID[Account]                `json:"-" bun:"account_id,notnull,pk"`
	Account                *Account                   `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	SimpleFinLinkId        ID[SimpleFinLink]          `json:"simpleFinLinkId" bun:"simplefin_link_id,notnull,nullzero"`
	SimpleFinLink          *SimpleFinLink             `json:"-" bun:"rel:belongs-to,join:simplefin_link_id=simplefin_link_id,join:account_id=account_id"`
	SimpleFinId            string                     `json:"simpleFinId" bun:"simplefin_id,notnull,nullzero"`
	ConnectionId           string                     `json:"connectionId" bun:"connection_id,notnull,nullzero"`
	Name                   string                     `json:"name" bun:"name,notnull,nullzero"`
	InstitutionName        string                     `json:"institutionName" bun:"institution_name,notnull,nullzero"`
	Currency               string                     `json:"currency" bun:"currency,notnull,nullzero"`
	Status                 SimpleFinBankAccountStatus `json:"status" bun:"status,notnull,nullzero"`
	CurrentBalance         int64                      `json:"currentBalance" bun:"current_balance,notnull"`
	AvailableBalance       int64                      `json:"availableBalance" bun:"available_balance,notnull"`
	CreatedAt              time.Time                  `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy              ID[User]                   `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser          *User                      `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
	UpdatedAt              time.Time                  `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	DeletedAt              *time.Time                 `json:"deletedAt,omitempty" bun:"deleted_at"`
}

func (SimpleFinBankAccount) IdentityPrefix() string {
	return "sbac"
}

var (
	_ bun.BeforeAppendModelHook = (*SimpleFinBankAccount)(nil)
)

func (o *SimpleFinBankAccount) BeforeAppendModel(
	ctx context.Context,
	query bun.Query,
) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.SimpleFinBankAccountId.IsZero() {
			o.SimpleFinBankAccountId = NewID[SimpleFinBankAccount]()
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
