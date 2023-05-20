package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (c *Controller) getAccountSettings(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	settings, err := repo.GetSettings(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve account settings")
	}

	return ctx.JSON(http.StatusOK, settings)
}

func (c *Controller) deleteAccount(ctx echo.Context) error {
	// TODO Implement a way to delete account data.
	return echo.NewHTTPError(http.StatusNotImplemented, "account deletion not yet implemented")
}
