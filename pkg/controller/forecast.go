package controller

import (
	"context"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/forecast"
	"github.com/monetr/monetr/pkg/models"
)

func (c *Controller) handleForecasting(p iris.Party) {
	p.Post("/{bankAccountId:uint64}/forecast/spending", c.postForecastNewSpending)
	p.Post("/{bankAccountId:uint64}/forecast/next_funding", c.postForecastNextFunding)
}

func (c *Controller) postForecastNewSpending(ctx iris.Context) {
	var request struct {
		FundingScheduleId uint64              `json:"fundingScheduleId"`
		SpendingType      models.SpendingType `json:"spendingType"`
		TargetAmount      int64               `json:"targetAmount"`
		CurrentAmount     int64               `json:"currentAmount"`
		NextRecurrence    time.Time           `json:"nextRecurrence"`
		RecurrenceRule    *models.Rule        `json:"recurrenceRule"`
	}
	if err := ctx.ReadJSON(&request); err != nil {
		c.invalidJson(ctx)
		return
	}

	if request.TargetAmount <= 0 {
		c.badRequest(ctx, "Target amount must be greater than 0")
		return
	}
	if request.CurrentAmount < 0 {
		c.badRequest(ctx, "Current amount cannot be less than 0")
		return
	}
	if request.FundingScheduleId == 0 {
		c.badRequest(ctx, "Funding schedule must be specified")
		return
	}
	if request.SpendingType == models.SpendingTypeExpense && request.RecurrenceRule == nil {
		c.badRequest(ctx, "Expense spending must have a recurrence rule")
		return
	}
	if request.SpendingType == models.SpendingTypeGoal && request.RecurrenceRule != nil {
		c.badRequest(ctx, "Goal spending must not have a recurrence rule")
		return
	}

	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "Must specify a valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedules, err := repo.GetFundingSchedules(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not retrieve funding schedules")
		return
	}

	afterForecast := forecast.NewForecaster([]models.Spending{
		{
			FundingScheduleId: request.FundingScheduleId,
			SpendingType:      request.SpendingType,
			TargetAmount:      request.TargetAmount,
			CurrentAmount:     request.CurrentAmount,
			NextRecurrence:    request.NextRecurrence,
			RecurrenceRule:    request.RecurrenceRule,
			SpendingId:        0, // Make sure this ID does not overlap with any real spending objects.
		},
	}, fundingSchedules)

	end := request.NextRecurrence.AddDate(2, 0, 0)
	timezone := c.mustGetTimezone(ctx)
	timeout, cancel := context.WithTimeout(c.getContext(ctx), 25*time.Second)
	defer cancel()
	ctx.JSON(map[string]interface{}{
		"estimatedCost": afterForecast.GetAverageContribution(
			timeout,
			request.NextRecurrence.AddDate(0, 0, -1),
			end,
			timezone,
		),
	})
}

func (c *Controller) postForecastNextFunding(ctx iris.Context) {
	var request struct {
		FundingScheduleId uint64 `json:"fundingScheduleId"`
	}
	if err := ctx.ReadJSON(&request); err != nil {
		c.invalidJson(ctx)
		return
	}

	if request.FundingScheduleId == 0 {
		c.badRequest(ctx, "Funding schedule must be specified")
		return
	}

	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "Must specify a valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedule, err := repo.GetFundingSchedule(c.getContext(ctx), bankAccountId, request.FundingScheduleId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not retrieve funding schedule")
		return
	}

	spending, err := repo.GetSpendingByFundingSchedule(c.getContext(ctx), bankAccountId, request.FundingScheduleId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not retrieve spending for forecast")
		return
	}

	fundingForecast := forecast.NewForecaster(
		spending,
		[]models.FundingSchedule{
			*fundingSchedule,
		},
	)
	timezone := c.mustGetTimezone(ctx)
	timeout, cancel := context.WithTimeout(c.getContext(ctx), 25*time.Second)
	defer cancel()
	ctx.JSON(map[string]interface{}{
		"nextContribution": fundingForecast.GetNextContribution(
			timeout,
			time.Now(),
			request.FundingScheduleId,
			timezone,
		),
	})
}
