package controller

// SubscriptionNotActiveError is returned to the client whenever they attempt to make an API call to an endpoint that
// requires an active subscription.
type SubscriptionNotActiveError struct {
	// Will include a message indicating that the user's subscription is not active. Will always be returned with a 402
	// status code.
	Error string `json:"error" example:"subscription is not active"`
}

// MalformedJSONError is returned to the client when the request body is not valid JSON or cannot be properly parsed.
type MalformedJSONError struct {
	// Will include a message indicating that the request body is not valid JSON.
	Error string `json:"error" example:"malformed json"`
}

type ApiError struct {
	Error string `json:"error" example:"something went wrong on our end"`
}

type InvalidBankAccountIdError struct {
	Error string `json:"error" example:"invalid bank account Id provided"`
}
