package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

func (c *Controller) expensesController(p iris.Party) {
	p.Get("/", func(ctx *context.Context) {

	})

	p.Post("/", func(ctx *context.Context) {

	})

	p.Put("/{expenseId:uint64}", func(ctx *context.Context) {

	})

	p.Delete("/{expenseId:uint64}", func(ctx *context.Context) {

	})
}
