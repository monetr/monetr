package swag

import (
	"github.com/monetr/monetr/pkg/models"
)

type MeResponse struct {
	// User has all of the details about the currently authenticated user.
	User models.User `json:"user"`
	// IsSetup indicates whether or not the user has any links configured at all. If this is false then the user should
	// be prompted to set up a new link.
	IsSetup bool `json:"isSetup"`
	// IsActive is an indicator of whether or not a user's subscription is active. This field is always present and is
	// true if billing is not configured on the instance at all. It is only ever false if the server has billing
	// configured _and_ the user's subscription has expired or has never been set up.
	IsActive bool `json:"isActive"`
	// HasSubscription is another billing indicator, this field is only present on the response when the server does
	// have billing enabled. It indicates whether or not the user does have a subscription, but not whether that
	// subscription is active. This can be true even if the subscription has expired.
	HasSubscription *bool `json:"hasSubscription" extensions:"x-nullable"`
	// NextURL is only present when the user needs to be directed somewhere by the backend. Currently, it is leveraged
	// for billing purposes. But in the future it can be used to direct the user through an on-boarding flow or
	// something similar.
	NextURL *string `json:"nextUrl" extensions:"x-nullable"`
}
