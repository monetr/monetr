package swag

import (
	"github.com/monetr/monetr/pkg/models"
	"time"
)

type CreateLinkRequest struct {
	// Specify the institution name for the manual link. When created by the UI this will default to `Manual` and then
	// any bank account's created that are manual will automatically be created for this link. Technically a link can
	// be created with any name at the moment, but this is meant to be a basic initial implementation for now.
	InstitutionName string `json:"institutionName" example:"Manual" validate:"required"`
}

type UpdateLinkRequest struct {
	// A custom name for the link's institution that can be set by the end user. Once a link is created the institution
	// name cannot be changed directly. But the custom institution name can.
	CustomInstitutionName string `json:"customInstitutionName" example:"US Bank"`
}

type LinkResponse struct {
	// Our unique identifier for a link. This is globally unique across all accounts.
	LinkId uint64 `json:"linkId" example:"1245"`
	// The type of link this object is. This indicates whether or not bank accounts within this link are managed
	// manually by an end user, or managed automatically by a Plaid integration.
	// * 0 - `Unknown`: This would indicate an error state with the link.
	// * 1 - `Plaid`: This link is automatically managed by the Plaid integration.
	// * 2 - `Manual`: This link is managed manually by the end user.
	LinkType models.LinkType `json:"linkType" example:"2" enums:"1,2"`
	// Status of a link, this is used with the Plaid integration to determine whether or not a link has been completely
	// setup. If a link is in an Unknown or Pending state then the link has not had it's transactions retrieved yet from
	// Plaid. Unknown might indicate there is a problem with the link itself.
	// * 0 - `Unknown`: Indicates the link is not setup or ready to use, might also indicate there is a problem with the link.
	// * 1 - `Pending`: The link is not ready to use and is being setup by the Plaid integration.
	// * 2 - `Setup`: The link is ready to use. This is the default state for manual links.
	// * 3 - `Error`: The link is in an error state, this can happen if the Plaid link is experiencing problems.
	LinkStatus models.LinkStatus `json:"linkStatus" example:"2" enums:"0,1,2"`
	// If the link error is due to a problem on Plaid's side, then an error code will be included here to help display
	// helpful messages on the frontend to the user.
	ErrorCode *string `json:"errorCode" extensions:"x-nullable" example:"NO_ACCOUNTS"`
	// Our internal Id for an institution. This is just an abstraction layer on top of Plaid's institution Id but would
	// allow us to associate institutions with multiple integrations in the future. It is also meant to keep Plaid Id's
	// away from the client's view as much as possible.
	InstitutionId *uint64 `json:"institutionId" example:"5328" extensions:"x-nullable"`
	UpdateLinkRequest
	// The institution name for this link. With the Plaid integration, each link represents a single bank account login
	// with an institution. So if you link your monetr account with your U.S. Bank account you would have one link for
	// U.S. Bank, and your accounts with that bank would show as bank account's under that link. Manual links are
	// intended to be a single "Manual" link per account, with any manually managed bank accounts underneath it.
	InstitutionName string `json:"institutionName" example:"U.S. Bank"`
	// The unique user Id of the user who created this link. If that user is deleted then any links that that user
	// created are also deleted. In an "ownership" sort of way. If a user links their bank account and has that shared
	// with someone else; but then the user deletes their user. We do not want the other user to still have access to
	// the linked bank account.
	CreatedByUserId uint64 `json:"createdByUserId" example:"94832"`
	// The timestamp the link was created in UTC. This cannot be modified by the end user.
	CreatedAt time.Time `json:"createdAt" example:"2021-05-18T00:26:10.873089Z"`
	// The user who last updated this link. Currently this field is not maintained well and should not be trusted.
	UpdatedByUserId *uint64 `json:"updatedByUserId" example:"89547" extensions:"x-nullable"`
	// The last time this link was updated. Currently this field is not really maintained, eventually this timestamp
	// will indicate the last time a sync occurred between monetr and Plaid. Manual links don't change this field at
	// all. **OLD**
	UpdatedAt time.Time `json:"updatedAt" example:"2021-05-21T05:24:12.958309Z"`
	// The last time transactions were successfully retrieved for this link. This date does not indicate the most recent
	// transaction retrieved, simply the most recent attempt to retrieve transactions that was successful.
	LastSuccessfulUpdate *time.Time `json:"lastSuccessfulUpdate" example:"2021-05-21T05:24:12.958309Z" extensions:"x-nullable"`
}
