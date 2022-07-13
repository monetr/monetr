package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func (c *Controller) handleAccounts(p router.Party) {
	p.Get("/settings", c.getAccountSettings)
	p.Delete("/", c.deleteAccount)
}

func (c *Controller) getAccountSettings(ctx iris.Context) {
	repo := c.mustGetAuthenticatedRepository(ctx)

	settings, err := repo.GetSettings(c.getContext(ctx))
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve account settings")
		return
	}

	ctx.JSON(settings)
}

func (c *Controller) deleteAccount(ctx iris.Context) {
	// TODO Implement a way to delete account data.
}
