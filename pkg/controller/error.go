package controller

// SubscriptionNotActiveError is returned to the client whenever they attempt to make an API call to an endpoint that
// requires an active subscription.
type SubscriptionNotActiveError struct {
	Error string `json:"error" example:"subscription is not active"`
}

type ApiError struct {
	Error string `json:"error" example:"something went wrong on our end"`
}

type InvalidBankAccountIdError struct {
	Error string `json:"error" example:"invalid bank account Id provided"`
}
