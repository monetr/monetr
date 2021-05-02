package swag

type CreateCheckoutSessionRequest struct {
	// PriceId represents the Id of the price object for the subscription. Price objects are associated with a single
	// product. So a price represents both how much is being paid, and what is being paid for.
	PriceId uint64 `json:"priceId"`
}

type CreateCheckoutSessionResponse struct {
	// The value returned from stripe once a checkout session has been created. This is used on the frontend for the
	// user to checkout and pay for their chosen plan.
	SessionId string `json:"sessionId"`
}
