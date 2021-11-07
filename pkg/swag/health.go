package swag

import "time"

type HealthResponse struct {
	// Indicates whether or not the current API process handling the request can communicate with the PostgreSQL
	// database.
	DBHealthy bool `json:"dbHealthy" example:"true"`

	// This will always be true. If the API is not healthy then an error is returned to the client or the request will
	// simply not be served.
	ApiHealthy bool `json:"apiHealthy" example:"true"`

	// The Git SHA code for the commit of the deployed REST API.
	Revision string `json:"revision" example:"c1becbe0654c4d8fde74c349ef7596983361eb7c"`

	// Release is only present when a deployment was run for a specific tag. This is only found in acceptance and
	// production.
	Release *string `json:"release" extensions:"x-nullable" example:"v0.3.9"`

	// The time the current REST API executable was built. Typically, when the container build was initiated.
	BuildTime time.Time `json:"buildTime" example:"2021-11-07T12:11:10Z"`

	// The current time on the server that handled the request. This is always in UTC.
	ServerTime time.Time `json:"serverTime" example:"2021-11-07T12:11:10Z"`
}
