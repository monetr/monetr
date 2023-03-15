package models

import "time"

type PlaidSync struct {
	tableName string `pg:"plaid_syncs"`

	PlaidSyncID uint64     `json:"plaidSyncId" pg:"plaid_sync_id,notnull,pk,type:'bigserial'"`
	PlaidLinkID uint64     `json:"-" pg:"plaid_link_id,notnull,on_delete:CASCADE,type:'bigint'"`
	PlaidLink   *PlaidLink `json:"-" pg:"rel:has-one"`
	Timestamp   time.Time  `json:"timestamp" pg:"timestamp,notnull"`
	Trigger     string     `json:"trigger" pg:"trigger,notnull"`
	NextCursor  string     `json:"-" pg:"cursor,notnull"`
	Added       int        `json:"added" pg:"added,notnull,use_zero"`
	Modified    int        `json:"modified" pg:"modified,notnull,use_zero"`
	Removed     int        `json:"removed" pg:"removed,notnull,use_zero"`
}
