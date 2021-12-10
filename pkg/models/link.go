package models

import (
	"time"

	"github.com/uptrace/bun"
)

type LinkStatus uint8

//go:generate stringer -type=LinkStatus -output=link.strings.go
const (
	LinkStatusUnknown           LinkStatus = 0
	LinkStatusPending           LinkStatus = 1
	LinkStatusSetup             LinkStatus = 2
	LinkStatusError             LinkStatus = 3
	LinkStatusPendingExpiration LinkStatus = 4
	LinkStatusRevoked           LinkStatus = 5
)

type Link struct {
	bun.BaseModel `bun:"links"`

	LinkId                uint64     `json:"linkId" bun:"link_id,notnull,pk"`
	AccountId             uint64     `json:"-" bun:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account               *Account   `json:"-" bun:"rel:has-one,join:account_id=account_id"`
	LinkType              LinkType   `json:"linkType" bun:"link_type,notnull"`
	PlaidLinkId           *uint64    `json:"-" bun:"plaid_link_id,on_delete:SET NULL"`
	PlaidLink             *PlaidLink `json:"-" bun:"rel:has-one,join:plaid_link_id=plaid_link_id"`
	PlaidInstitutionId    *string    `json:"plaidInstitutionId" bun:"plaid_institution_id"`
	LinkStatus            LinkStatus `json:"linkStatus" bun:"link_status,notnull,default:0"`
	ErrorCode             *string    `json:"errorCode,omitempty" bun:"error_code"`
	ExpirationDate        *time.Time `json:"expirationDate" bun:"expiration_date"`
	InstitutionName       string     `json:"institutionName" bun:"institution_name"`
	CustomInstitutionName string     `json:"customInstitutionName,omitempty" bun:"custom_institution_name"`
	CreatedAt             time.Time  `json:"createdAt" bun:"created_at,notnull"`
	CreatedByUserId       uint64     `json:"createdByUserId" bun:"created_by_user_id,notnull,on_delete:CASCADE"`
	CreatedByUser         *User      `json:"-,omitempty" bun:"rel:has-one,join:created_by_user_id=user_id"`
	UpdatedAt             time.Time  `json:"updatedAt" bun:"updated_at,notnull"`
	LastSuccessfulUpdate  *time.Time `json:"lastSuccessfulUpdate" bun:"last_successful_update"`

	BankAccounts []BankAccount `json:"-" bun:"rel:has-many,join:link_id=link_id"`
}
