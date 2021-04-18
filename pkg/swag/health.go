package swag

type HealthResponse struct {
	// Indicates whether or not the current API process handling the request can communicate with the PostgreSQL
	// database.
	DBHealthy bool `json:"dbHealthy"`

	// This will always be true. If the API is not healthy then an error is returned to the client or the request will
	// simply not be served.
	ApiHealthy bool `json:"apiHealthy"`
}
