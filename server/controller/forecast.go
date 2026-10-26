package controller

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/forecast"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/schemas"
	"github.com/teambition/rrule-go"
)

func (c *Controller) getForecast(ctx *echo.Context) error {
	now := time.Now()
	log := c.getLog(ctx)
	var endDate time.Time
	end := ctx.QueryParam("end")
	if strings.TrimSpace(end) == "" {
		log.Log(c.getContext(ctx), logging.LevelTrace, "no end date specified for forecast, will default to 31 days")
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

func (c *Controller) postForecastNewSpending(ctx *echo.Context) error {
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

	return ctx.JSON(http.StatusOK, map[string]any{
		"estimatedCost": result,
	})
}

func (c *Controller) postForecastNextFunding(ctx *echo.Context) error {
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

	return ctx.JSON(http.StatusOK, map[string]any{
		"nextContribution": result,
	})
}

func (c *Controller) queryInitialFundingForRecurringTransaction(ctx *echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "Must specify a valid bank account Id")
	}

	var request *struct {
		TransactionRecurringId models.ID[models.TransactionRecurring] `json:"transactionRecurringId"`
		FundingScheduleId      models.ID[models.FundingSchedule]      `json:"fundingScheduleId"`
	}
	request, err = parse(c, ctx, request, schemas.ForecastInitialFunding)
	if err != nil {
		return err
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	timezone := c.mustGetTimezone(ctx)

	recurring, err := repo.GetTransactionRecurringById(
		c.getContext(ctx),
		bankAccountId,
		request.TransactionRecurringId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve recurring transaction")
	}

	funding, err := repo.GetFundingSchedule(
		c.getContext(ctx),
		bankAccountId,
		request.FundingScheduleId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve funding schedule")
	}

	now := c.Clock.Now().In(timezone)
	rule := recurring.RuleSet.InTimezone(timezone)
	// Start from the last time this would have been spent, everything allocated
	// after that is for the next occurrence. Which is what this expense should
	// have allocated today.
	start := rule.Before(now, true)

	// The forecaster uses the next recurrence of the funding schedule as is when
	// we are before it, which skips every contribution between start and now. So
	// clear it and the forecaster will walk the rule instead.
	fundingSchedule := *funding
	fundingSchedule.NextRecurrence = time.Time{}
	fundingSchedule.RuleSet = funding.RuleSet.InTimezone(timezone)
	options := fundingSchedule.RuleSet.GetRRule().OrigOptions
	// Moving DTSTART back would use up the count before today
	if options.Count > 0 {
		return c.badRequest(ctx, "Funding schedules with a count are not supported")
	}

	// If there are no contributions before start then the forecaster panics, this
	// happens when the funding schedule was created after the last time this was
	// spent. So move DTSTART back like the funding schedule has always existed.
	// It has to move back by whole periods so the paydays stay on the same days,
	// an every other week rule moved back 7 days would get paid on the wrong
	// weeks. This is done in the account's timezone, otherwise moving back across
	// DST lands on the day before.
	for fundingSchedule.RuleSet.Before(start, false).IsZero() {
		dtstart := fundingSchedule.RuleSet.GetDTStart()
		switch options.Freq {
		case rrule.DAILY, rrule.WEEKLY:
			dtstart = dtstart.AddDate(0, 0, -7*options.Interval)
		case rrule.MONTHLY, rrule.YEARLY:
			// Whole years so we never land on a day that doesn't exist, Apr 30 minus
			// 2 months is Feb 30 which Go turns into Mar 2.
			dtstart = dtstart.AddDate(-options.Interval, 0, 0)
		default:
			return c.badRequest(ctx, "Funding schedule frequency is not supported")
		}
		fundingSchedule.RuleSet.DTStart(dtstart)
	}

	forecaster := forecast.NewForecaster(
		c.getLog(ctx),
		[]models.Spending{
			{
				FundingScheduleId: request.FundingScheduleId,
				SpendingType:      models.SpendingTypeExpense,
				TargetAmount:      recurring.LastAmount,
				CurrentAmount:     0,
				RuleSet:           rule,
				NextRecurrence:    start,
			},
		},
		[]FundingSchedule{
			fundingSchedule,
		},
	)

	forecast, err := forecaster.GetForecast(
		c.getContext(ctx),
		start,
		now,
		timezone,
	)
	if err == context.DeadlineExceeded {
		return c.returnError(ctx, http.StatusRequestTimeout, "timeout forecasting")
	} else if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to forecast")
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		// Everything contributed since the last time this was spent, like the
		// expense had been there the whole time
		"currentAmount": forecast.EndingBalance,
	})
}
