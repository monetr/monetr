package models

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type LunchFlowLinkStatus string

const (
	// LunchFlowLinkStatusPending is for when the lunch flow link has been created
	// but has not been fully setup. The link is not yet associated with any
	// actual monetr links and is not being used to feed data into the application
	// yet. This status may be cleaned up after some period of time.
	LunchFlowLinkStatusPending LunchFlowLinkStatus = "pending"
	// LunchFlowLinkStatusActive is the status used once the lunch flow link has
	// been associated with the monetr link. This is used to filter automated
	// syncing with the background jobs and this link will be picked up for data
	// syncing periodically.
	LunchFlowLinkStatusActive LunchFlowLinkStatus = "active"
	// LunchFlowLinkStatusDeactivated is when the link has been manually
	// deactivated or is pending removal. This status will prevent the link from
	// being picked up by automated background jobs.
	LunchFlowLinkStatusDeactivated LunchFlowLinkStatus = "deactivated"
	// LunchFlowLinkStatusError is used for when the background process has made
	// multiple attempts to sync the data for the link but has only encountered
	// errors when attempting to do so. To prevent the background jobs from
	// continuing to attempt to sync this link, the link is moved to an error
	// status. The user can move the link back to an active status manually via
	// the user interface or API.
	// TODO This status is not propagated automatically yet.
	LunchFlowLinkStatusError LunchFlowLinkStatus = "error"
)

type LunchFlowLink struct {
	tableName string `pg:"lunch_flow_links"`

	LunchFlowLinkId      ID[LunchFlowLink]   `json:"lunchFlowLinkId" pg:"lunch_flow_link_id,notnull,pk"`
	AccountId            ID[Account]         `json:"-" pg:"account_id,pk,notnull"`
	Account              *Account            `json:"-" pg:"rel:has-one"`
	SecretId             ID[Secret]          `json:"-" pg:"secret_id,notnull"`
	Secret               *Secret             `json:"-" pg:"rel:has-one"`
	ApiUrl               string              `json:"apiUrl" pg:"api_url,notnull"`
	Status               LunchFlowLinkStatus `json:"status" pg:"status,notnull"`
	LastManualSync       *time.Time          `json:"lastManualSync" pg:"last_manual_sync"`
	LastSuccessfulUpdate *time.Time          `json:"lastSuccessfulUpdate" pg:"last_successful_update"`
	LastAttemptedUpdate  *time.Time          `json:"lastAttemptedUpdate" pg:"last_attempted_update"`
	UpdatedAt            time.Time           `json:"updatedAt" pg:"updated_at,notnull"`
	CreatedAt            time.Time           `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy            ID[User]            `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser        *User               `json:"-" pg:"rel:has-one,fk:created_by"`
	DeletedAt            *time.Time          `json:"deletedAt,omitempty" pg:"deleted_at"`
}

func (LunchFlowLink) IdentityPrefix() string {
	return "lfx"
}

var (
	_ pg.BeforeInsertHook = (*LunchFlowLink)(nil)
)

func (o *LunchFlowLink) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.LunchFlowLinkId.IsZero() {
		o.LunchFlowLinkId = NewID[LunchFlowLink]()
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
