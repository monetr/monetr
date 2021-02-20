package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

func (c *Controller) fundingSchedulesController(p iris.Party) {
	p.Get("/", func(ctx *context.Context) {

	})

	p.Post("/", func(ctx *context.Context) {

	})

	p.Put("/{fundingScheduleId:uint64}", func(ctx *context.Context) {

	})

	p.Delete("/{fundingScheduleId:uint64}", func(ctx *context.Context) {

	})
}
