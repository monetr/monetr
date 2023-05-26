package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/icons"
)

func (c *Controller) searchIcon(ctx echo.Context) error {
	if !icons.GetIconsEnabled() {
		return c.notFound(ctx, "icons are not enabled")
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := ctx.Bind(&body); err != nil {
		return c.invalidJson(ctx)
	}

	if body.Name == "" {
		return c.badRequest(ctx, "must provide a name to search icons for")
	}

	icon, err := icons.SearchIcon(body.Name)
	if err != nil || icon == nil {
		return ctx.NoContent(http.StatusNoContent)
	}

	return ctx.JSON(http.StatusOK, icon)
}
