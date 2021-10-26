package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/monetr/monetr/pkg/models"
)

func (c *Controller) handleFundingSchedules(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/funding_schedules", c.getFundingSchedules)
	p.Get("/{bankAccountId:uint64}/funding_schedules/stats", c.getFundingScheduleStats)
	p.Post("/{bankAccountId:uint64}/funding_schedules", c.postFundingSchedules)
	p.Put("/{bankAccountId:uint64}/funding_schedules/{fundingScheduleId:uint64}", c.putFundingSchedules)
	p.Delete("/{bankAccountId:uint64}/funding_schedules/{fundingScheduleId:uint64}", c.deleteFundingSchedules)
}

// List Funding Schedules
// @Summary List Funding Schedules
// @id list-funding-schedules
// @tags Funding Schedules
// @description List all of the funding schedule's for the current bank account.
// @Security ApiKeyAuth
// @Produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Router /bank_accounts/{bankAccountId}/funding_schedules [get]
// @Success 200 {array} models.FundingSchedule
// @Failure 400 {object} InvalidBankAccountIdError Invalid Bank Account ID.
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getFundingSchedules(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedules, err := repo.GetFundingSchedules(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve funding schedules")
		return
	}

	if fundingSchedules == nil {
		fundingSchedules = make([]models.FundingSchedule, 0)
	}

	ctx.JSON(fundingSchedules)
}

// Get Funding Stats
// @Summary Get Funding Stats
// @id get-funding-status
// @tags Funding Schedules
// @description Retrieve information about how much spending objects will receive on the next funding schedule.
// @Security ApiKeyAuth
// @Produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Router /bank_accounts/{bankAccountId}/funding_schedules/stats [get]
// @Success 200 {array} repository.FundingStats
// @Failure 400 {object} InvalidBankAccountIdError Invalid Bank Account ID.
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getFundingScheduleStats(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	stats, err := repo.GetFundingStats(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve funding schedules")
		return
	}

	ctx.JSON(stats)
}

// Create Funding Schedule
// @Summary Create Funding Schedule
// @id create-funding-schedule
// @tags Funding Schedules
// @security ApiKeyAuth
// @accept json
// @produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Param fundingSchedule body models.FundingSchedule true "New Funding Schedule"
// @Router /bank_accounts/{bankAccountId}/funding_schedules [post]
// @Success 200 {object} models.FundingSchedule
// @Failure 400 {object} ApiError "Malformed JSON or invalid RRule."
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError "Failed to persist data."
func (c *Controller) postFundingSchedules(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	var fundingSchedule models.FundingSchedule
	if err := ctx.ReadJSON(&fundingSchedule); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	fundingSchedule.FundingScheduleId = 0 // Make sure we create a new funding schedule.
	fundingSchedule.Name = strings.TrimSpace(fundingSchedule.Name)
	fundingSchedule.Description = strings.TrimSpace(fundingSchedule.Description)
	fundingSchedule.BankAccountId = bankAccountId

	if fundingSchedule.Name == "" {
		c.returnError(ctx, http.StatusBadRequest, "funding schedule must have a name")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	var err error
	// Set the next occurrence based on the provided rule.

	// If the next occurrence is not specified then assume that the rule is relative to now. If it is specified though
	// then do nothing, and the subsequent occurrence will be calculated relative to the provided date.
	if (time.Time{}).Equal(fundingSchedule.NextOccurrence) {
		fundingSchedule.NextOccurrence, err = c.midnightInLocal(ctx, fundingSchedule.Rule.After(time.Now(), false))
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to determine next occurrence")
			return
		}
	}

	// It has never occurred so this needs to be nil.
	fundingSchedule.LastOccurrence = nil

	if err = repo.CreateFundingSchedule(c.getContext(ctx), &fundingSchedule); err != nil {
		c.wrapPgError(ctx, err, "failed to create funding schedule")
		return
	}

	ctx.JSON(fundingSchedule)
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
