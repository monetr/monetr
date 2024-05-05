package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (c *Controller) deleteAccount(ctx echo.Context) error {
	// TODO Implement a way to delete account data.
	return echo.NewHTTPError(http.StatusNotImplemented, "account deletion not yet implemented")
}
