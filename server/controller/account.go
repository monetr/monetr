package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func (*Controller) deleteAccount(_ *echo.Context) error {
	// TODO Implement a way to delete account data.
	return echo.NewHTTPError(http.StatusNotImplemented, "account deletion not yet implemented")
}
