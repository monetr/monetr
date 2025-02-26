package main

import (
	"github.com/go-pg/pg/v10"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/middleware"
	"github.com/monetr/monetr/server/repository"
)

// RegisterAPIKeyMiddleware registers the API key authentication middleware with the Echo application.
// This middleware will check for an API key in the X-API-Key header and authenticate the request if
// a valid API key is provided.
func RegisterAPIKeyMiddleware(app *echo.Echo, repo repository.APIKeyRepository, db pg.DBI) {
	// Add the API key middleware to the global middleware chain
	// This middleware will check for an API key and set the user ID in the context if found
	app.Use(middleware.APIKeyAuthentication(repo, db))
}
