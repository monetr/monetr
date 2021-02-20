package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

func (c *Controller) transactionsController(p iris.Party) {
	p.Get("/", func(ctx *context.Context) {

	})

	p.Post("/", func(ctx *context.Context) {

	})

	p.Put("/{transactionId:uint64}", func(ctx *context.Context) {

	})

	p.Delete("/{transactionId:uint64}", func(ctx *context.Context) {

	})
}
