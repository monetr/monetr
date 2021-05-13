package swag

import "time"

type HealthResponse struct {
	// Indicates whether or not the current API process handling the request can communicate with the PostgreSQL
	// database.
	DBHealthy bool `json:"dbHealthy"`

	// This will always be true. If the API is not healthy then an error is returned to the client or the request will
	// simply not be served.
	ApiHealthy bool `json:"apiHealthy"`

	// The Git SHA code for the commit of the deployed REST API.
	Revision string `json:"revision"`

	// Release is only present when a deployment was run for a specific tag. This is only found in acceptance and
	// production.
	Release *string `json:"release"`

	// The time the current REST API executable was built. Typically when the container build was initiated.
	BuildTime time.Time `json:"buildTime"`
}
