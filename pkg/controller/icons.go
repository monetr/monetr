package controller

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/icons"
)

func (c *Controller) iconsController(p iris.Party) {
	if icons.GetIconsEnabled() {
		// DEPRECATED Use the POST method.
		p.Get("/search", c.searchIcon)
		p.Post("/search", c.searchIcon)
	}
}

func (c *Controller) searchIcon(ctx iris.Context) {
	var name string
	switch ctx.Method() {
	case "GET":
		name = ctx.URLParam("name")
	case "POST":
		var body struct {
			Name string `json:"name"`
		}
		if err := ctx.ReadJSON(&body); err != nil {
			c.invalidJson(ctx)
			return
		}
		name = body.Name
	}

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
