package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/icons"
	"github.com/monetr/monetr/server/schemas"
)

func (c *Controller) searchIcon(ctx *echo.Context) error {
	if !icons.GetIconsEnabled() {
		return c.notFound(ctx, "icons are not enabled")
	}

	var body *struct {
		Name string `json:"name"`
	}
	body, err := parse(c, ctx, body, schemas.IconSearch)
	if err != nil {
		return err
	}

	icon, err := icons.SearchIcon(body.Name)
	if err != nil || icon == nil {
		return ctx.NoContent(http.StatusNoContent)
	}

	return ctx.JSON(http.StatusOK, icon)
}
