package swag

type CreateCheckoutSessionRequest struct {
	// Specify a specific Stripe Price ID to be used when creating the checkout session. If this is left blank then
	// the default price will be used for the checkout session.
	PriceId *string `json:"priceId" example:"price_1JFQFuI4uGGnwpgwquHOo34s" extensions:"x-nullable"`
	// The path that the user should be returned to if they exit the checkout session.
	CancelPath *string `json:"cancelPath"`
}

type CreateCheckoutSessionResponse struct {
	// The value returned from stripe once a checkout session has been created. This is used on the frontend for the
	// user to checkout and pay for their chosen plan.
	SessionId string `json:"sessionId"`
}

type CreatePortalSessionResponse struct {
	// The URL returned by Stripe for the customer's billing portal.
	URL string `json:"url"`
}

// AfterCheckoutResponse is returned from the after checkout endpoint. It will either indicate that the subscription
// has been completely and properly setup, or that the subscription was not activated and the user still does not have
// full application access.
type AfterCheckoutResponse struct {
	// Message is included if there is a problem. Right now this happens if the checkout session is completed but the
	// subscription associated with that checkout session is not active.
	//
	// **NOTE:** This field is not included if the subscription is active and the after checkout is successful.
	Message  string `json:"message,omitempty" example:"Subscription is not active" extensions:"x-nullable"`
	// IsActive is used to indicate whether the user's subscription is not properly activated. On the UI this is
	// propagated to the redux store to allow access to other application routes. If this is false then the subscription
	// is not active and API calls to endpoints requiring payment will still fail.
	IsActive bool   `json:"isActive"`
	// NextURL is used to direct the user to a specific page after their checkout has been completed and verified. This
	// should be followed by the web UI. Right now, successful checkouts will redirect to `/` which will prompt the user
	// to either configure a Plaid link, or will present them with their budgeting data if there already is some.
	NextURL  string `json:"nextUrl"`
}
