package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"net/http"
)

func (c *Controller) handleFundingSchedules(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/funding_schedules", c.getFundingSchedules)
	p.Post("/{bankAccountId:uint64}/funding_schedules", c.postFundingSchedules)
	p.Put("/{bankAccountId:uint64}/funding_schedules/{fundingScheduleId:uint64}", c.putFundingSchedules)
	p.Delete("/{bankAccountId:uint64}/funding_schedules/{fundingScheduleId:uint64}", c.deleteFundingSchedules)
}

// List Funding Schedules
// @id list-funding-schedules
// @description List all of the funding schedule's for the current bank account.
// @Router /bank_accounts/{bankAccountId}/funding_schedules [get]
// @Success 200 {array} models.FundingSchedule
func (c *Controller) getFundingSchedules(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

}

func (c *Controller) postFundingSchedules(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

}

func (c *Controller) putFundingSchedules(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

}

func (c *Controller) deleteFundingSchedules(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

}
