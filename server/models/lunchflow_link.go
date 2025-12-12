package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type LunchFlowLink struct {
	tableName string `pg:"lunchflow_links"`

	LunchFlowLinkId      ID[LunchFlowLink] `json:"lunchFlowLinkId" pg:"lunchflow_link_id,notnull,pk"`
	AccountId            ID[Account]       `json:"-" pg:"account_id,notnull"`
	Account              *Account          `json:"-" pg:"rel:has-one"`
	SecretId             ID[Secret]        `json:"-" pg:"secret_id,notnull"`
	Secret               *Secret           `json:"-" pg:"rel:has-one"`
	ApiUrl               string            `json:"apiUrl" pg:"api_url,notnull"`
	LastManualSync       *time.Time        `json:"lastManualSync" pg:"last_manual_sync"`
	LastSuccessfulUpdate *time.Time        `json:"lastSuccessfulUpdate" pg:"last_successful_update"`
	LastAttemptedUpdate  *time.Time        `json:"lastAttemptedUpdate" pg:"last_attempted_update"`
	UpdatedAt            time.Time         `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedAt            time.Time         `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy            ID[User]          `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser        *User             `json:"-" pg:"rel:has-one,fk:created_by"`
	DeletedAt            *time.Time        `json:"deletedAt" pg:"deleted_at"`
}

func (LunchFlowLink) IdentityPrefix() string {
	return "lfx"
}

var (
	_ pg.BeforeInsertHook = (*LunchFlowLink)(nil)
)

func (o *LunchFlowLink) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.LunchFlowLinkId.IsZero() {
		o.LunchFlowLinkId = NewID(o)
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
