package controller

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/forecast"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) getForecast(ctx echo.Context) error {
	now := time.Now()
	log := c.getLog(ctx)
	var endDate time.Time
	end := ctx.QueryParam("end")
	if strings.TrimSpace(end) == "" {
		log.Trace("no end date specified for forecast, will default to 31 days")
		endDate = now.AddDate(0, 0, 90).UTC()
	} else {
		var err error
		endDate, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return c.badRequest(ctx, "invalid end time provided, end time must be in RFC3339 format")
		}
	}

	if endDate.Before(now) {
		return c.badRequest(ctx, "invalid end time provided, end time must be in the future")
	}
	if endDate.After(now.AddDate(0, 12, 0)) {
		return c.badRequest(ctx, "you are not allowed for forecast more than 12 months into the future")
	}

	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedules, err := repo.GetFundingSchedules(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retrieve funding schedules")
	}

	spending, err := repo.GetSpending(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retreive spending")
	}

	forecaster := forecast.NewForecaster(
		log,
		spending,
		fundingSchedules,
	)

	timezone := c.mustGetTimezone(ctx)
	timeout, cancel := context.WithTimeout(c.getContext(ctx), 15*time.Second)
	defer cancel()

	result, err := forecaster.GetForecast(
		timeout,
		now,
		endDate,
		timezone,
	)
	if err == context.DeadlineExceeded {
		return c.returnError(ctx, http.StatusRequestTimeout, "timeout forecasting")
	} else if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to forecast")
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *Controller) postForecastNewSpending(ctx echo.Context) error {
	var request struct {
		FundingScheduleId ID[FundingSchedule] `json:"fundingScheduleId"`
		SpendingType      SpendingType        `json:"spendingType"`
		TargetAmount      int64               `json:"targetAmount"`
		CurrentAmount     int64               `json:"currentAmount"`
		NextRecurrence    time.Time           `json:"nextRecurrence"`
		RuleSet           *RuleSet            `json:"recurrenceRule"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	if request.TargetAmount <= 0 {
		return c.badRequest(ctx, "Target amount must be greater than 0")
	}
	if request.CurrentAmount < 0 {
		return c.badRequest(ctx, "Current amount cannot be less than 0")
	}
	if request.FundingScheduleId.IsZero() {
		return c.badRequest(ctx, "Funding schedule must be specified")
	}
	if request.SpendingType == SpendingTypeExpense && request.RuleSet == nil {
		return c.badRequest(ctx, "Expense spending must have a recurrence rule")
	}
	if request.SpendingType == SpendingTypeGoal && request.RuleSet != nil {
		return c.badRequest(ctx, "Goal spending must not have a recurrence rule")
	}

	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedules, err := repo.GetFundingSchedules(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retrieve funding schedules")
	}
	log := c.getLog(ctx)

	afterForecast := forecast.NewForecaster(
		log,
		[]Spending{
			{
				FundingScheduleId: request.FundingScheduleId,
				SpendingType:      request.SpendingType,
				TargetAmount:      request.TargetAmount,
				CurrentAmount:     request.CurrentAmount,
				NextRecurrence:    request.NextRecurrence,
				RuleSet:           request.RuleSet,
				SpendingId:        "spnd_forecast", // Make sure this ID does not overlap with any real spending objects.
			},
		},
		fundingSchedules,
	)

	end := request.NextRecurrence.AddDate(2, 0, 0)
	timezone := c.mustGetTimezone(ctx)
	timeout, cancel := context.WithTimeout(c.getContext(ctx), 25*time.Second)
	defer cancel()

	result, err := afterForecast.GetAverageContribution(
		timeout,
		request.NextRecurrence.AddDate(0, 0, -1),
		end,
		timezone,
	)
	if err == context.DeadlineExceeded {
		return c.returnError(ctx, http.StatusRequestTimeout, "timeout forecasting")
	} else if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to forecast")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"estimatedCost": result,
	})
}

func (c *Controller) postForecastNextFunding(ctx echo.Context) error {
	var request struct {
		FundingScheduleId ID[FundingSchedule] `json:"fundingScheduleId"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	if request.FundingScheduleId.IsZero() {
		return c.badRequest(ctx, "Funding schedule must be specified")
	}

	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedule, err := repo.GetFundingSchedule(c.getContext(ctx), bankAccountId, request.FundingScheduleId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retrieve funding schedule")
	}

	spending, err := repo.GetSpendingByFundingSchedule(c.getContext(ctx), bankAccountId, request.FundingScheduleId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retrieve spending for forecast")
	}
	log := c.getLog(ctx)

	fundingForecast := forecast.NewForecaster(
		log,
		spending,
		[]FundingSchedule{
			*fundingSchedule,
		},
	)
	timezone := c.mustGetTimezone(ctx)
	timeout, cancel := context.WithTimeout(c.getContext(ctx), 25*time.Second)
	defer cancel()
	result, err := fundingForecast.GetNextContribution(
		timeout,
		c.Clock.Now(),
		request.FundingScheduleId,
		timezone,
	)
	if err == context.DeadlineExceeded {
		return c.returnError(ctx, http.StatusRequestTimeout, "timeout forecasting")
	} else if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to forecast")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"nextContribution": result,
	})
}
