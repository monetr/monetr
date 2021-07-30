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
