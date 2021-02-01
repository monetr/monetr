package controller

import (
	"net/http"

	"github.com/kataras/iris/v12/context"
)

func (c *Controller) meEndpoint(ctx *context.Context) {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusForbidden, "cannot retrieve user details")
		return
	}

	user, err := repo.GetMe()
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "cannot retrieve user details")
		return
	}

	isSetup, err := repo.GetIsSetup()
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not determine if account is setup")
		return
	}

	ctx.JSON(map[string]interface{}{
		"user":    user,
		"isSetup": isSetup,
	})
}
