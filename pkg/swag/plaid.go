package swag

type PlaidNewLinkTokenResponse struct {
	// The link token that will be used for the end user to authenticate to their bank using plaid. These tokens do
	// expire. They are also specific to a single environment. See: https://plaid.com/docs/api/tokens/#linktokencreate
	LinkToken string `json:"linkToken" example:"link-environment-6da2c37f-6aa0...."`
}

type PlaidLinkLimitError struct {
	// Error will include a message about how the user has reached their limit for Plaid links.
	Error string `json:"error" example:"max number of Plaid links already reached"`
}

type PlaidTokenCallbackResponse struct {
	Success bool `json:"success"`
	// LinkId will always be included in a successful response. It can be used when webhooks are enabled to wait for the
	// initial transactions to be retrieved.
	LinkId uint64 `json:"linkId"`
	// If webhooks are not enabled then a job Id is returned with the response. This job Id can also be used to check
	// for initial transactions being retrieved.
	JobId *string `json:"jobId" extensions:"x-nullable"`
}

type UpdatePlaidTokenCallbackRequest struct {
	LinkId      uint64 `json:"linkId"`
	PublicToken string `json:"publicToken"`
}

type NewPlaidTokenCallbackRequest struct {
	PublicToken     string   `json:"publicToken"`
	InstitutionId   string   `json:"institutionId" example:"ins_117212"`
	InstitutionName string   `json:"institutionName" example:"Navy Federal Credit Union"`
	AccountIds      []string `json:"accountIds" example:"KEdQjMo39lFwXKqKLlqEt6R3AgBWW1C6l8vDn,r3DVlexNymfJkgZgonZeSQ4n5Koqqjtyrwvkp"`
}
