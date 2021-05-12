package swag

type PlaidNewLinkTokenResponse struct {
	// The link token that will be used for the end user to authenticate to their bank using plaid. These tokens do
	// expire. They are also specific to a single environment. See: https://plaid.com/docs/api/tokens/#linktokencreate
	LinkToken string `json:"linkToken" example:"link-environment-6da2c37f-6aa0...."`
}

type PlaidTokenCallbackResponse struct {
	Success bool    `json:"success"`
	// LinkId will always be included in a successful response. It can be used when webhooks are enabled to wait for the
	// initial transactions to be retrieved.
	LinkId  uint64  `json:"linkId"`
	// If webhooks are not enabled then a job Id is returned with the response. This job Id can also be used to check
	// for initial transactions being retrieved.
	JobId   *string `json:"jobId"`
}
