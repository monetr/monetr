package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
)

func (c *Controller) handleFundingSchedules(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/funding_schedules", func(ctx *context.Context) {

	})

	p.Post("/{bankAccountId:uint64}/funding_schedules", func(ctx *context.Context) {

	})

	p.Put("/{bankAccountId:uint64}/funding_schedules/{fundingScheduleId:uint64}", func(ctx *context.Context) {

	})

	p.Delete("/{bankAccountId:uint64}/funding_schedules/{fundingScheduleId:uint64}", func(ctx *context.Context) {

	})
}
