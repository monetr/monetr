package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type PlaidSync struct {
	bun.BaseModel `bun:"table:plaid_syncs,alias:plaid_sync"`

	PlaidSyncId ID[PlaidSync] `json:"plaidSyncId" bun:"plaid_sync_id,notnull,pk"`
	AccountId   ID[Account]   `json:"-" bun:"account_id,notnull,pk"`
	Account     *Account      `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	PlaidLinkId ID[PlaidLink] `json:"-" bun:"plaid_link_id,notnull,nullzero"`
	PlaidLink   *PlaidLink    `json:"-" bun:"rel:belongs-to,join:plaid_link_id=plaid_link_id"`
	Timestamp   time.Time     `json:"timestamp" bun:"timestamp,notnull,nullzero"`
	Trigger     string        `json:"trigger" bun:"trigger,notnull,nullzero"`
	NextCursor  string        `json:"-" bun:"cursor,notnull,nullzero"`
	Added       int           `json:"added" bun:"added,notnull"`
	Modified    int           `json:"modified" bun:"modified,notnull"`
	Removed     int           `json:"removed" bun:"removed,notnull"`
}

func (PlaidSync) IdentityPrefix() string {
	return "psyn"
}

var (
	_ bun.BeforeAppendModelHook = (*PlaidSync)(nil)
)

func (o *PlaidSync) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.PlaidSyncId.IsZero() {
			o.PlaidSyncId = NewID[PlaidSync]()
		}
	}

	return nil
}
