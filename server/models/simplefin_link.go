package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type SimpleFinLinkStatus string

const (
	// SimpleFinLinkStatusPending is for when the SimpleFIN link has been created
	// but has not been fully setup yet. Pending links get cleaned up after some
	// period of time along with their secret, which is a working access URL.
	SimpleFinLinkStatusPending SimpleFinLinkStatus = "pending"
	// SimpleFinLinkStatusActive is the status used once the SimpleFIN link has
	// been associated with a monetr link. Only active links are picked up by the
	// background sync.
	SimpleFinLinkStatusActive SimpleFinLinkStatus = "active"
	// SimpleFinLinkStatusDeactivated is when the link has been manually
	// deactivated or is pending removal.
	SimpleFinLinkStatusDeactivated SimpleFinLinkStatus = "deactivated"
	// SimpleFinLinkStatusError is used when SimpleFIN tells us the access URL is
	// no longer usable, like when the token was revoked or the user needs to pay
	// for the Bridge. We stop syncing the link automatically at that point
	// because every request counts against the user's daily quota. A successful
	// manual sync will move the link back to active.
	SimpleFinLinkStatusError SimpleFinLinkStatus = "error"
)

// SimpleFinLinkError is a single item from the errlist SimpleFIN returns with
// every request. SimpleFIN wants these shown to the user so we keep the last
// set of them on the link.
type SimpleFinLinkError struct {
	Code         string  `json:"code"`
	Message      string  `json:"message"`
	ConnectionId *string `json:"connectionId"`
	SimpleFinId  *string `json:"simpleFinId"`
}

type SimpleFinLink struct {
	bun.BaseModel `bun:"table:simplefin_links,alias:simplefin_link"`

	SimpleFinLinkId      ID[SimpleFinLink]    `json:"simpleFinLinkId" bun:"simplefin_link_id,notnull,pk"`
	AccountId            ID[Account]          `json:"-" bun:"account_id,notnull,pk"`
	Account              *Account             `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	SecretId             ID[Secret]           `json:"-" bun:"secret_id,notnull,nullzero"`
	Secret               *Secret              `json:"-" bun:"rel:belongs-to,join:secret_id=secret_id,join:account_id=account_id"`
	Name                 string               `json:"name" bun:"name,notnull,nullzero"`
	ApiUrl               string               `json:"apiUrl" bun:"api_url,notnull,nullzero"`
	Status               SimpleFinLinkStatus  `json:"status" bun:"status,notnull,nullzero"`
	LastErrors           []SimpleFinLinkError `json:"lastErrors" bun:"last_errors,type:jsonb,nullzero"`
	LastManualSync       *time.Time           `json:"lastManualSync" bun:"last_manual_sync"`
	LastSuccessfulUpdate *time.Time           `json:"lastSuccessfulUpdate" bun:"last_successful_update"`
	LastAttemptedUpdate  *time.Time           `json:"lastAttemptedUpdate" bun:"last_attempted_update"`
	CreatedAt            time.Time            `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy            ID[User]             `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser        *User                `json:"-" bun:"rel:belongs-to,join:created_by=user_id"`
	UpdatedAt            time.Time            `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	DeletedAt            *time.Time           `json:"deletedAt,omitempty" bun:"deleted_at"`
}

func (SimpleFinLink) IdentityPrefix() string {
	return "sfx"
}

var (
	_ bun.BeforeAppendModelHook = (*SimpleFinLink)(nil)
)

func (o *SimpleFinLink) BeforeAppendModel(
	ctx context.Context,
	query bun.Query,
) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.SimpleFinLinkId.IsZero() {
			o.SimpleFinLinkId = NewID[SimpleFinLink]()
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
