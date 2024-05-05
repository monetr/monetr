package models

import "time"

type Link struct {
	tableName string `pg:"links"`

	LinkId          ID[Link]       `json:"linkId" pg:"link_id,notnull,pk"`
	AccountId       ID[Account]    `json:"-" pg:"account_id,notnull,pk"`
	Account         *Account       `json:"-" pg:"rel:has-one"`
	LinkType        LinkType       `json:"linkType" pg:"link_type,notnull"`
	PlaidLinkId     *ID[PlaidLink] `json:"-" pg:"plaid_link_id"`
	PlaidLink       *PlaidLink     `json:"plaidLink,omitempty" pg:"rel:has-one"`
	InstitutionName string         `json:"institutionName" pg:"institution_name"`
	Description     *string        `json:"description" pg:"description"`
	CreatedAt       time.Time      `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy       ID[User]       `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser   *User          `json:"-,omitempty" pg:"rel:has-one,fk:created_by"`
	UpdatedAt       time.Time      `json:"updatedAt" pg:"updated_at,notnull"`
	DeletedAt       *time.Time     `json:"deletedAt" pg:"deleted_at"`
}

func (o Link) IdentityPrefix() string {
	return "link"
}
