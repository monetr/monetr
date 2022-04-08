package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func (c *Controller) handleAccounts(p router.Party) {
	p.Delete("/", c.deleteAccount)
}

func (c *Controller) deleteAccount(ctx iris.Context) {
	// TODO Implement a way to delete account data.
}
