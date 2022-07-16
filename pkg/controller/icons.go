package controller

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/icons"
)

func (c *Controller) iconsController(p iris.Party) {
	if icons.GetIconsEnabled() {
		p.Get("/search", c.searchIcon)
	}
}

func (c *Controller) searchIcon(ctx iris.Context) {
	name := ctx.URLParam("name")
	if name == "" {
		c.badRequest(ctx, "must provide a name to search icons for")
		return
	}

	icon, err := icons.SearchIcon(name)
	if err != nil || icon == nil {
		ctx.StatusCode(http.StatusNoContent)
		return
	}

	ctx.JSON(icon)
}
