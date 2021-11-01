package models

import "time"

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
	tableName string `pg:"links"`

	LinkId                uint64     `json:"linkId" pg:"link_id,notnull,pk,type:'bigserial'"`
	AccountId             uint64     `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account               *Account   `json:"-" pg:"rel:has-one"`
	LinkType              LinkType   `json:"linkType" pg:"link_type,notnull"`
	PlaidLinkId           *uint64    `json:"-" pg:"plaid_link_id,on_delete:SET NULL"`
	PlaidLink             *PlaidLink `json:"-" pg:"rel:has-one"`
	PlaidInstitutionId    *string    `json:"plaidInstitutionId" pg:"plaid_institution_id"`
	LinkStatus            LinkStatus `json:"linkStatus" pg:"link_status,notnull,default:0"`
	ErrorCode             *string    `json:"errorCode,omitempty" pg:"error_code"`
	ExpirationDate        *time.Time `json:"expirationDate" pg:"expiration_date"`
	InstitutionName       string     `json:"institutionName" pg:"institution_name"`
	CustomInstitutionName string     `json:"customInstitutionName,omitempty" pg:"custom_institution_name"`
	CreatedAt             time.Time  `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId       uint64     `json:"createdByUserId" pg:"created_by_user_id,notnull,on_delete:CASCADE"`
	CreatedByUser         *User      `json:"-,omitempty" pg:"rel:has-one,fk:created_by_user_id"`
	UpdatedAt             time.Time  `json:"updatedAt" pg:"updated_at,notnull"`
	UpdatedByUserId       *uint64    `json:"updatedByUserId" pg:"updated_by_user_id,on_delete:SET NULL"`
	UpdatedByUser         *User      `json:"updatedByUser,omitempty" pg:"rel:has-one,fk:updated_by_user_id"`
	LastSuccessfulUpdate  *time.Time `json:"lastSuccessfulUpdate" pg:"last_successful_update"`

	BankAccounts []BankAccount `json:"-" pg:"rel:has-many"`
}
