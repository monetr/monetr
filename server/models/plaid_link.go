package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type PlaidLinkStatus string

const (
	PlaidLinkStatusUnknown           PlaidLinkStatus = "unknown"
	PlaidLinkStatusPending           PlaidLinkStatus = "pending"
	PlaidLinkStatusSetup             PlaidLinkStatus = "setup"
	PlaidLinkStatusError             PlaidLinkStatus = "error"
	PlaidLinkStatusPendingExpiration PlaidLinkStatus = "pending_expiration"
	PlaidLinkStatusRevoked           PlaidLinkStatus = "revoked"
	PlaidLinkStatusDeactivated       PlaidLinkStatus = "deactivated"
)

type PlaidLink struct {
	bun.BaseModel `bun:"table:plaid_links,alias:plaid_link"`

	PlaidLinkId          ID[PlaidLink]   `json:"-" bun:"plaid_link_id,notnull,pk"`
	AccountId            ID[Account]     `json:"-" bun:"account_id,notnull,nullzero"`
	Account              *Account        `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	SecretId             ID[Secret]      `json:"-" bun:"secret_id,nullzero"`
	Secret               *Secret         `json:"-" bun:"rel:belongs-to,join:secret_id=secret_id,join:account_id=account_id"`
	PlaidId              string          `json:"-" bun:"item_id,unique,notnull,nullzero"`
	Products             []string        `json:"products" bun:"products,array,nullzero"`
	Status               PlaidLinkStatus `json:"status" bun:"status,notnull,nullzero"`
	ErrorCode            *string         `json:"errorCode,omitempty" bun:"error_code"`
	ExpirationDate       *time.Time      `json:"expirationDate" bun:"expiration_date"`
	NewAccountsAvailable bool            `json:"newAccountsAvailable" bun:"new_accounts_available"`
	WebhookUrl           string          `json:"-" bun:"webhook_url,nullzero"`
	InstitutionId        string          `json:"institutionId" bun:"institution_id,notnull,nullzero"`
	InstitutionName      string          `json:"institutionName" bun:"institution_name,nullzero"`
	LastManualSync       *time.Time      `json:"lastManualSync" bun:"last_manual_sync"`
	LastSuccessfulUpdate *time.Time      `json:"lastSuccessfulUpdate" bun:"last_successful_update"`
	LastAttemptedUpdate  *time.Time      `json:"lastAttemptedUpdate" bun:"last_attempted_update"`
	LastAccountSync      *time.Time      `json:"lastAccountSync" bun:"last_account_sync"`
	UpdatedAt            time.Time       `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	CreatedAt            time.Time       `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy            ID[User]        `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser        *User           `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
	DeletedAt            *time.Time      `json:"deletedAt" bun:"deleted_at"`
}

func (PlaidLink) IdentityPrefix() string {
	return "plx"
}

var (
	_ bun.BeforeAppendModelHook = (*PlaidLink)(nil)
)

func (o *PlaidLink) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.PlaidLinkId.IsZero() {
			o.PlaidLinkId = NewID[PlaidLink]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}

		if o.UpdatedAt.IsZero() {
			o.UpdatedAt = now
		}
	}

	return nil
}
