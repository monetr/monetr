package models

import "time"

type PlaidLinkStatus uint8

//go:generate stringer -type=PlaidLinkStatus -output=plaid_link.strings.go
const (
	PlaidLinkStatusUnknown           PlaidLinkStatus = 0
	PlaidLinkStatusPending           PlaidLinkStatus = 1
	PlaidLinkStatusSetup             PlaidLinkStatus = 2
	PlaidLinkStatusError             PlaidLinkStatus = 3
	PlaidLinkStatusPendingExpiration PlaidLinkStatus = 4
	PlaidLinkStatusRevoked           PlaidLinkStatus = 5
)

type PlaidLink struct {
	tableName string `pg:"plaid_links"`

	PlaidLinkId          uint64          `json:"-" pg:"plaid_link_id,notnull,pk,type:'bigserial'"`
	AccountId            uint64          `json:"-" pg:"account_id,notnull,type:'bigint'"`
	Account              *Account        `json:"-" pg:"rel:has-one"`
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
	UpdatedAt            time.Time       `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedAt            time.Time       `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId      uint64          `json:"createdByUserId" pg:"created_by_user_id,notnull"`
	CreatedByUser        *User           `json:"-" pg:"rel:has-one,fk:created_by_user_id"`
}
