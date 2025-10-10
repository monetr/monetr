package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type PlaidLinkStatus uint8

//go:generate go run golang.org/x/tools/cmd/stringer@v0.38.0 -type=PlaidLinkStatus -output=plaid_link.strings.go
const (
	PlaidLinkStatusUnknown           PlaidLinkStatus = 0
	PlaidLinkStatusPending           PlaidLinkStatus = 1
	PlaidLinkStatusSetup             PlaidLinkStatus = 2
	PlaidLinkStatusError             PlaidLinkStatus = 3
	PlaidLinkStatusPendingExpiration PlaidLinkStatus = 4
	PlaidLinkStatusRevoked           PlaidLinkStatus = 5
	PlaidLinkStatusDeactivated       PlaidLinkStatus = 6
)

type PlaidLink struct {
	tableName string `pg:"plaid_links"`

	PlaidLinkId          ID[PlaidLink]   `json:"-" pg:"plaid_link_id,notnull,pk,type:'bigserial'"`
	AccountId            ID[Account]     `json:"-" pg:"account_id,notnull,type:'bigint'"`
	Account              *Account        `json:"-" pg:"rel:has-one"`
	SecretId             ID[Secret]      `json:"-" pg:"secret_id,type:'bigint'"`
	Secret               *Secret         `json:"-" pg:"rel:has-one"`
	PlaidId              string          `json:"-" pg:"item_id,unique,notnull"`
	Products             []string        `json:"products" pg:"products,type:'text[]'"`
	Status               PlaidLinkStatus `json:"status" pg:"status,notnull,default:0"`
	ErrorCode            *string         `json:"errorCode,omitempty" pg:"error_code"`
	ExpirationDate       *time.Time      `json:"expirationDate" pg:"expiration_date"`
	NewAccountsAvailable bool            `json:"newAccountsAvailable" pg:"new_accounts_available,use_zero"`
	WebhookUrl           string          `json:"-" pg:"webhook_url"`
	InstitutionId        string          `json:"institutionId" pg:"institution_id,notnull"`
	InstitutionName      string          `json:"institutionName" pg:"institution_name"`
	LastManualSync       *time.Time      `json:"lastManualSync" pg:"last_manual_sync"`
	LastSuccessfulUpdate *time.Time      `json:"lastSuccessfulUpdate" pg:"last_successful_update"`
	LastAttemptedUpdate  *time.Time      `json:"lastAttemptedUpdate" pg:"last_attempted_update"`
	LastAccountSync      *time.Time      `json:"lastAccountSync" pg:"last_account_sync"`
	UpdatedAt            time.Time       `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedAt            time.Time       `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy            ID[User]        `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser        *User           `json:"-" pg:"rel:has-one,fk:created_by"`
	DeletedAt            *time.Time      `json:"deletedAt" pg:"deleted_at"`
}

func (PlaidLink) IdentityPrefix() string {
	return "plx"
}

var (
	_ pg.BeforeInsertHook = (*PlaidLink)(nil)
)

func (o *PlaidLink) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.PlaidLinkId.IsZero() {
		o.PlaidLinkId = NewID(o)
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
