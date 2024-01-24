package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (c *Controller) postTellerLink(ctx echo.Context) error {
	if !c.configuration.Teller.GetEnabled() {
		return c.returnError(ctx, http.StatusNotAcceptable, "Teller is not enabled on this server.")
	}

	return ctx.NoContent(http.StatusOK)
}
