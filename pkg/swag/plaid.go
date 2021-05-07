package swag

type PlaidTokenCallbackResponse struct {
	Success bool    `json:"success"`
	// LinkId will always be included in a successful response. It can be used when webhooks are enabled to wait for the
	// initial transactions to be retrieved.
	LinkId  uint64  `json:"linkId"`
	// If webhooks are not enabled then a job Id is returned with the response. This job Id can also be used to check
	// for initial transactions being retrieved.
	JobId   *string `json:"jobId"`
}
