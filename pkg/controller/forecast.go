package controller

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/forecast"
	"github.com/monetr/monetr/pkg/models"
)

func (c *Controller) postForecastNewSpending(ctx echo.Context) error {
	var request struct {
		FundingScheduleId uint64              `json:"fundingScheduleId"`
		SpendingType      models.SpendingType `json:"spendingType"`
		TargetAmount      int64               `json:"targetAmount"`
		CurrentAmount     int64               `json:"currentAmount"`
		NextRecurrence    time.Time           `json:"nextRecurrence"`
		RecurrenceRule    *models.Rule        `json:"recurrenceRule"`
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
	if request.FundingScheduleId == 0 {
		return c.badRequest(ctx, "Funding schedule must be specified")
	}
	if request.SpendingType == models.SpendingTypeExpense && request.RecurrenceRule == nil {
		return c.badRequest(ctx, "Expense spending must have a recurrence rule")
	}
	if request.SpendingType == models.SpendingTypeGoal && request.RecurrenceRule != nil {
		return c.badRequest(ctx, "Goal spending must not have a recurrence rule")
	}

	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
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
		[]models.Spending{
			{
				FundingScheduleId: request.FundingScheduleId,
				SpendingType:      request.SpendingType,
				TargetAmount:      request.TargetAmount,
				CurrentAmount:     request.CurrentAmount,
				NextRecurrence:    request.NextRecurrence,
				RecurrenceRule:    request.RecurrenceRule,
				SpendingId:        0, // Make sure this ID does not overlap with any real spending objects.
			},
		},
		fundingSchedules,
	)

	end := request.NextRecurrence.AddDate(2, 0, 0)
	timezone := c.mustGetTimezone(ctx)
	timeout, cancel := context.WithTimeout(c.getContext(ctx), 25*time.Second)
	defer cancel()

	result, err := func() (result int64, err error) {
		defer func() {
			switch panicResult := recover().(type) {
			case error:
				err = panicResult
				return
			}
		}()
		result = afterForecast.GetAverageContribution(
			timeout,
			request.NextRecurrence.AddDate(0, 0, -1),
			end,
			timezone,
		)
		return result, nil
	}()
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
		FundingScheduleId uint64 `json:"fundingScheduleId"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	if request.FundingScheduleId == 0 {
		return c.badRequest(ctx, "Funding schedule must be specified")
	}

	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
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
		[]models.FundingSchedule{
			*fundingSchedule,
		},
	)
	timezone := c.mustGetTimezone(ctx)
	timeout, cancel := context.WithTimeout(c.getContext(ctx), 25*time.Second)
	defer cancel()
	result, err := func() (result int64, err error) {
		defer func() {
			switch panicResult := recover().(type) {
			case error:
				err = panicResult
				return
			}
		}()
		result = fundingForecast.GetNextContribution(
			timeout,
			time.Now(),
			request.FundingScheduleId,
			timezone,
		)
		return result, nil
	}()
	if err == context.DeadlineExceeded {
		return c.returnError(ctx, http.StatusRequestTimeout, "timeout forecasting")
	} else if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to forecast")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"nextContribution": result,
	})
}
