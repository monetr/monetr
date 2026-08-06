package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
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
	bun.BaseModel `bun:"table:lunch_flow_links,alias:lunch_flow_link"`

	LunchFlowLinkId      ID[LunchFlowLink]   `json:"lunchFlowLinkId" bun:"lunch_flow_link_id,notnull,pk"`
	AccountId            ID[Account]         `json:"-" bun:"account_id,pk,notnull"`
	Account              *Account            `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	SecretId             ID[Secret]          `json:"-" bun:"secret_id,notnull,nullzero"`
	Secret               *Secret             `json:"-" bun:"rel:belongs-to,join:secret_id=secret_id,join:account_id=account_id"`
	Name                 string              `json:"name" bun:"name,notnull,nullzero"`
	ApiUrl               string              `json:"apiUrl" bun:"api_url,notnull,nullzero"`
	Status               LunchFlowLinkStatus `json:"status" bun:"status,notnull,nullzero"`
	LastManualSync       *time.Time          `json:"lastManualSync" bun:"last_manual_sync"`
	LastSuccessfulUpdate *time.Time          `json:"lastSuccessfulUpdate" bun:"last_successful_update"`
	LastAttemptedUpdate  *time.Time          `json:"lastAttemptedUpdate" bun:"last_attempted_update"`
	UpdatedAt            time.Time           `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	CreatedAt            time.Time           `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy            ID[User]            `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser        *User               `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
	DeletedAt            *time.Time          `json:"deletedAt,omitempty" bun:"deleted_at"`
}

func (LunchFlowLink) IdentityPrefix() string {
	return "lfx"
}

var (
	_ bun.BeforeAppendModelHook = (*LunchFlowLink)(nil)
)

func (o *LunchFlowLink) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
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
	}

	return nil
}
