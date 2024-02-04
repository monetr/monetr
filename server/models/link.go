package models

import "time"

type Link struct {
	tableName string `pg:"links"`

	LinkId          uint64      `json:"linkId" pg:"link_id,notnull,pk,type:'bigserial'"`
	AccountId       uint64      `json:"-" pg:"account_id,notnull,pk,on_delete:CASCADE,type:'bigint'"`
	Account         *Account    `json:"-" pg:"rel:has-one"`
	LinkType        LinkType    `json:"linkType" pg:"link_type,notnull"`
	PlaidLinkId     *uint64     `json:"-" pg:"plaid_link_id"`
	PlaidLink       *PlaidLink  `json:"plaidLink,omitempty" pg:"rel:has-one"`
	TellerLinkId    *uint64     `json:"-" pg:"teller_link_id"`
	TellerLink      *TellerLink `json:"tellerLink,omitempty" pg:"rel:has-one"`
	InstitutionName string      `json:"institutionName" pg:"institution_name"`
	Description     *string     `json:"description" pg:"description"`
	CreatedAt       time.Time   `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId uint64      `json:"createdByUserId" pg:"created_by_user_id,notnull"`
	CreatedByUser   *User       `json:"-,omitempty" pg:"rel:has-one,fk:created_by_user_id"`
	UpdatedAt       time.Time   `json:"updatedAt" pg:"updated_at,notnull"`
	DeletedAt       *time.Time  `json:"deletedAt" pg:"deleted_at"`
}
