package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (c *Controller) getFiles(ctx echo.Context) error {
	repo := c.mustGetAuthenticatedRepository(ctx)

	files, err := repo.GetFiles(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to list files")
	}

	return ctx.JSON(http.StatusOK, files)
}
