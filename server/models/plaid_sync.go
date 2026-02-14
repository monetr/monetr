package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type PlaidSync struct {
	tableName string `pg:"plaid_syncs"`

	PlaidSyncId ID[PlaidSync] `json:"plaidSyncId" pg:"plaid_sync_id,notnull,pk"`
	AccountId   ID[Account]   `json:"-" pg:"account_id,notnull,pk"`
	Account     *Account      `json:"-" pg:"rel:has-one"`
	PlaidLinkId ID[PlaidLink] `json:"-" pg:"plaid_link_id,notnull"`
	PlaidLink   *PlaidLink    `json:"-" pg:"rel:has-one"`
	Timestamp   time.Time     `json:"timestamp" pg:"timestamp,notnull"`
	Trigger     string        `json:"trigger" pg:"trigger,notnull"`
	NextCursor  string        `json:"-" pg:"cursor,notnull"`
	Added       int           `json:"added" pg:"added,notnull,use_zero"`
	Modified    int           `json:"modified" pg:"modified,notnull,use_zero"`
	Removed     int           `json:"removed" pg:"removed,notnull,use_zero"`
}

func (PlaidSync) IdentityPrefix() string {
	return "psyn"
}

var (
	_ pg.BeforeInsertHook = (*PlaidSync)(nil)
)

func (o *PlaidSync) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.PlaidSyncId.IsZero() {
		o.PlaidSyncId = NewID[PlaidSync]()
	}

	return ctx, nil
}
