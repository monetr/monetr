package models

import "time"

type TellerToken struct {
	tableName string `pg:"teller_tokens"`

	TellerTokenId   uint64      `json:"-" pg:"teller_token_id,notnull,pk,type:'bigserial'"`
	AccountId       uint64      `json:"-" pg:"account_id,notnull,type:'bigint',unique:per_account"`
	Account         *Account    `json:"-" pg:"rel:has-one"`
	TellerLinkId    uint64      `json:"-" pg:"teller_link_id,type:'bigint',unique:per_account"`
	TellerLink      *TellerLink `json:"-" pg:"rel:has-one"`
	KeyID           *string     `pg:"key_id"`
	Version         *string     `pg:"version"`
	AccessToken     string      `pg:"access_token,notnull"`
	UpdatedAt       time.Time   `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedAt       time.Time   `json:"createdAt" pg:"created_at,notnull"`
	CreatedByUserId uint64      `json:"createdByUserId" pg:"created_by_user_id,notnull"`
	CreatedByUser   *User       `json:"-" pg:"rel:has-one,fk:created_by_user_id"`
}
