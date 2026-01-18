package models

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/merge"
	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/pkg/errors"
)

type Link struct {
	tableName string `pg:"links"`

	LinkId          ID[Link]           `json:"linkId" pg:"link_id,notnull,pk"`
	AccountId       ID[Account]        `json:"-" pg:"account_id,notnull,pk"`
	Account         *Account           `json:"-" pg:"rel:has-one"`
	LinkType        LinkType           `json:"linkType" pg:"link_type,notnull"`
	PlaidLinkId     *ID[PlaidLink]     `json:"-" pg:"plaid_link_id"`
	PlaidLink       *PlaidLink         `json:"plaidLink,omitempty" pg:"rel:has-one"`
	LunchFlowLinkId *ID[LunchFlowLink] `json:"lunchFlowLinkId,omitempty" pg:"lunch_flow_link_id"`
	LunchFlowLink   *LunchFlowLink     `json:"lunchFlowLink,omitempty" pg:"rel:has-one"`
	InstitutionName string             `json:"institutionName" pg:"institution_name"`
	Description     *string            `json:"description" pg:"description"`
	CreatedAt       time.Time          `json:"createdAt" pg:"created_at,notnull"`
	CreatedBy       ID[User]           `json:"createdBy" pg:"created_by,notnull"`
	CreatedByUser   *User              `json:"-,omitempty" pg:"rel:has-one,fk:created_by"`
	UpdatedAt       time.Time          `json:"updatedAt" pg:"updated_at,notnull"`
	DeletedAt       *time.Time         `json:"deletedAt" pg:"deleted_at"`
}

func (o Link) IdentityPrefix() string {
	return "link"
}

var (
	_ pg.BeforeInsertHook = (*Link)(nil)
)

func (o *Link) BeforeInsert(ctx context.Context) (context.Context, error) {
	if o.LinkId.IsZero() {
		o.LinkId = NewID[Link]()
	}

	now := time.Now()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = now
	}

	return ctx, nil
}

// CreateValidators returns an array of validation rules that should be applied
// when creating a new instance of this object via the API. Only fields with
// validation rules should allow user input.
func (Link) CreateValidators() []*validation.KeyRules {
	return []*validation.KeyRules{
		validation.Key(
			"institutionName",
			validation.Required.Error("Institution name is required"),
			validation.Length(1, 300).Error("Institution name must be between 1 and 300 characters"),
		),
		validation.Key(
			"description",
			validation.Length(1, 300).Error("Description must be between 1 and 300 characters"),
		).Required(validators.Optional),
		validation.Key(
			"lunchFlowLinkId",
			ValidID[LunchFlowLink]().Error("Lunch Flow Link ID must be valid if provided"),
		),
	}
}

// UpdateValidator returns an array of validation rules that should be applied
// when updating this specific link object via an API call. Only the validated
// fields should be updated as well.
func (Link) UpdateValidator() []*validation.KeyRules {
	return []*validation.KeyRules{
		validation.Key(
			"institutionName",
			validation.Length(1, 300).Error("Institution name must be between 1 and 300 characters"),
		).Required(validators.Optional),
		validation.Key(
			"description",
			validation.Length(1, 300).Error("Description must be between 1 and 300 characters"),
		).Required(validators.Optional),
	}
}

// UnmarshalRequest consumes a request body and an array of validation rules in
// order to create an object that can be persisted to the database. For updates,
// this function should be called on the existing object that is already stored
// in the database. The provided validators should prevent key or sensitive
// fields from being overwritten by the client's request body. For creates, the
// initial object can be left blank; or default values can be specified ahead of
// calling this function in case some fields are omitted in the intial request.
func (o *Link) UnmarshalRequest(
	ctx context.Context,
	reader io.Reader,
	validators ...*validation.KeyRules,
) error {
	rawData := map[string]any{}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(&rawData); err != nil {
		return errors.WithStack(err)
	}

	if err := validation.ValidateWithContext(
		ctx,
		&rawData,
		validation.Map(
			validators...,
		),
	); err != nil {
		return err
	}

	if err := merge.Merge(
		o, rawData, merge.ErrorOnUnknownField,
	); err != nil {
		return errors.Wrap(err, "failed to merge patched data")
	}

	return nil
}
