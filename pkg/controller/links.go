package controller

import (
	"github.com/kataras/iris/v12/context"
	"net/http"
)

func (c *Controller) getLinksEndpoint(ctx *context.Context) {
	repo := c.mustGetAuthenticatedRepository(ctx)

	links, err := repo.GetLinks()
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve links")
		return
	}

	ctx.JSON(map[string]interface{}{
		"links": links,
	})
}
