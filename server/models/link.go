package models

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type Link struct {
	bun.BaseModel `bun:"table:links,alias:link"`

	LinkId          ID[Link]           `json:"linkId" bun:"link_id,notnull,pk"`
	AccountId       ID[Account]        `json:"-" bun:"account_id,notnull,pk"`
	Account         *Account           `json:"-" bun:"rel:belongs-to,join:account_id=account_id"`
	LinkType        LinkType           `json:"linkType" bun:"link_type,notnull,nullzero"`
	PlaidLinkId     *ID[PlaidLink]     `json:"-" bun:"plaid_link_id"`
	PlaidLink       *PlaidLink         `json:"plaidLink,omitempty" bun:"rel:belongs-to,join:plaid_link_id=plaid_link_id"`
	LunchFlowLinkId *ID[LunchFlowLink] `json:"lunchFlowLinkId,omitempty" bun:"lunch_flow_link_id"`
	LunchFlowLink   *LunchFlowLink     `json:"lunchFlowLink,omitempty" bun:"rel:belongs-to,join:lunch_flow_link_id=lunch_flow_link_id,join:account_id=account_id"`
	InstitutionName string             `json:"institutionName" bun:"institution_name,nullzero"`
	Description     *string            `json:"description" bun:"description"`
	CreatedAt       time.Time          `json:"createdAt" bun:"created_at,notnull,nullzero"`
	CreatedBy       ID[User]           `json:"createdBy" bun:"created_by,notnull,nullzero"`
	CreatedByUser   *User              `json:"-,omitempty" bun:"rel:belongs-to,join:created_by=user_id"`
	UpdatedAt       time.Time          `json:"updatedAt" bun:"updated_at,notnull,nullzero"`
	DeletedAt       *time.Time         `json:"deletedAt" bun:"deleted_at"`
}

func (Link) IdentityPrefix() string {
	return "link"
}

var (
	_ bun.BeforeAppendModelHook = (*Link)(nil)
)

func (o *Link) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		if o.LinkId.IsZero() {
			o.LinkId = NewID[Link]()
		}

		now := time.Now()
		if o.CreatedAt.IsZero() {
			o.CreatedAt = now
		}
	}

	return nil
}
