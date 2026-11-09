package schemas

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

type PostSimpleFinLinkRequest struct {
	Name       string `json:"name"`
	SetupToken string `json:"setupToken"`
}

type PostSimpleFinLinkSyncRequest struct {
	LinkId models.ID[models.Link] `json:"linkId"`
}

var (
	// PostSimpleFinLink is intended to be used alongside
	// [PostSimpleFinLinkRequest] in order to parse a non model request body.
	PostSimpleFinLink = validation.Map(
		validation.Key("name",
			Name(),
		).Required(Require),
		// Setup tokens are base64 so they can have +/= in them, don't use
		// is.UTFLetterNumeric here. The client does the real validation when it
		// decodes the token.
		validation.Key(
			"setupToken",
			validation.Required.Error("SimpleFIN setup token must be provided to setup a SimpleFIN link"),
			is.String,
			validation.Length(1, 2048).Error("SimpleFIN setup token must be between 1 and 2048 characters"),
		).Required(Require),
	)

	PostSimpleFinLinkSync = validation.Map(
		validation.Key(
			"linkId",
			validation.Required,
			ValidID[models.Link](),
		).Required(Require),
	)
)
