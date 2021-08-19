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

type LinkNotFoundError struct {
	// This error is returned when the user attempts to retrieve a link that does not exist or belong to their account.
	Error string `json:"error" example:"failed to retrieve link: record does not exist"`
}

type SpendingNotFoundError struct {
	// This error is returned when the user attempts to retrieve a spending object that does not exist or belong to their account.
	Error string `json:"error" example:"failed to retrieve spending: record does not exist"`
}

type InvalidLinkIdError struct {
	// Contains an error telling the user that they must provide a valid link Id for this request.
	Error string `json:"error" example:"must specify a link Id to retrieve"`
}

type InvalidBankAccountIdError struct {
	// Contains an error telling the user that they must provide a valid bank account Id for this request.
	Error string `json:"error" example:"invalid bank account Id provided"`
}
