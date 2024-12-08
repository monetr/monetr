package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/pkg/errors"
)

func (c *Controller) getFundingSchedules(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedules, err := repo.GetFundingSchedules(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve funding schedules")
	}

	if fundingSchedules == nil {
		fundingSchedules = make([]FundingSchedule, 0)
	}

	return ctx.JSON(http.StatusOK, fundingSchedules)
}

func (c *Controller) getFundingScheduleById(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fundingScheduleId, err := ParseID[FundingSchedule](ctx.Param("fundingScheduleId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid funding schedule Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedule, err := repo.GetFundingSchedule(
		c.getContext(ctx),
		bankAccountId,
		fundingScheduleId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve funding schedule")
	}

	return ctx.JSON(http.StatusOK, fundingSchedule)
}

func (c *Controller) postFundingSchedules(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	var fundingSchedule FundingSchedule
	if err := ctx.Bind(&fundingSchedule); err != nil {
		return c.invalidJson(ctx)
	}

	fundingSchedule.FundingScheduleId = "" // Make sure we create a new funding schedule.
	fundingSchedule.Name, err = c.cleanString(ctx, "Name", fundingSchedule.Name)
	if err != nil {
		return err
	}
	fundingSchedule.Description, err = c.cleanString(ctx, "Description", fundingSchedule.Description)
	if err != nil {
		return err
	}
	fundingSchedule.BankAccountId = bankAccountId

	if fundingSchedule.Name == "" {
		return c.badRequest(ctx, "funding schedule must have a name")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// Set the next occurrence based on the provided rule.

	// If the next occurrence is not specified then assume that the rule is relative to now. If it is specified though
	// then do nothing, and the subsequent occurrence will be calculated relative to the provided date.
	// We also calculate the next occurrence if the provided occurrence is in the past. This technically should not
	// happen via the UI. But it is currently possible for someone to select the current day in the UI. Which then gets
	// adjusted for midnight that day, which will always be in the past for the user.
	if fundingSchedule.NextRecurrence.IsZero() || c.Clock.Now().After(fundingSchedule.NextRecurrence) {
		fundingSchedule.CalculateNextOccurrence(
			c.getContext(ctx),
			c.Clock.Now(),
			c.mustGetTimezone(ctx),
		)
	} else {
		fundingSchedule.NextRecurrenceOriginal = fundingSchedule.NextRecurrence
	}

	// It has never occurred so this needs to be nil.
	fundingSchedule.LastRecurrence = nil

	if err = repo.CreateFundingSchedule(c.getContext(ctx), &fundingSchedule); err != nil {
		return c.wrapPgError(ctx, err, "failed to create funding schedule")
	}

	return ctx.JSON(http.StatusOK, fundingSchedule)
}

func (c *Controller) putFundingSchedules(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fundingScheduleId, err := ParseID[FundingSchedule](ctx.Param("fundingScheduleId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid funding schedule Id")
	}

	var request FundingSchedule
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	// Retrieve the existing funding schedule to make sure some fields cannot be overridden
	existingFundingSchedule, err := repo.GetFundingSchedule(
		c.getContext(ctx),
		bankAccountId,
		fundingScheduleId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to verify funding schedule exists")
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)

	// Don't allow these fields to be overwritten by an update API call.
	request.FundingScheduleId = fundingScheduleId
	request.BankAccountId = bankAccountId
	request.AccountId = existingFundingSchedule.AccountId
	request.LastRecurrence = existingFundingSchedule.LastRecurrence

	if request.Name == "" {
		return c.badRequest(ctx, "funding schedule must have a name")
	}

	if request.RuleSet == nil {
		return c.badRequest(ctx, "funding schedule must include a rule set")
	}

	if request.EstimatedDeposit != nil && *request.EstimatedDeposit < 0 {
		return c.badRequest(ctx, "estimated deposit must be greater than or equal to zero")
	}

	recalculateSpending := false
	// If the next occurrence changes then we need to recalulate spending.
	if !request.NextRecurrence.Equal(existingFundingSchedule.NextRecurrence) {
		// The user cannot override the next occurrence for a funding schedule and have it be in the past. If they set it to
		// be in the future then that is okay. The next time the funding schedule is processed it will be relative to that
		// next occurrence.
		if request.NextRecurrence.Before(c.Clock.Now()) {
			request.NextRecurrence = existingFundingSchedule.NextRecurrence
			request.NextRecurrenceOriginal = existingFundingSchedule.NextRecurrenceOriginal
		} else {
			request.NextRecurrenceOriginal = request.NextRecurrence
		}
		recalculateSpending = true
	}

	// If the recurrence rule has changed then we need to recalculate spending too.
	if request.RuleSet.String() != existingFundingSchedule.RuleSet.String() {
		recalculateSpending = true
	}

	updatedSpending := make([]Spending, 0)
	if recalculateSpending {
		crumbs.Debug(c.getContext(ctx), "Spending will be recalculated as part of this funding schedule update", map[string]interface{}{
			"bankAccountId":     bankAccountId,
			"fundingScheduleId": fundingScheduleId,
		})
		spending, err := repo.GetSpendingByFundingSchedule(c.getContext(ctx), bankAccountId, fundingScheduleId)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to retrieve existing spending objects for funding schedule")
		}

		if len(spending) > 0 {
			account, err := repo.GetAccount(c.getContext(ctx))
			if err != nil {
				return c.wrapPgError(ctx, err, "failed to retrieve account data to update funding schedule")
			}

			now := time.Now()
			for _, spend := range spending {
				if err := spend.CalculateNextContribution(c.getContext(ctx), account.Timezone, &request, now); err != nil {
					return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to recalculate spending object")
				}
			}

			if err = repo.UpdateSpending(c.getContext(ctx), bankAccountId, spending); err != nil {
				return c.wrapPgError(ctx, err, "failed to update spending objects for updated funding schedule")
			}
			updatedSpending = spending
		}
	}

	if err = repo.UpdateFundingSchedule(c.getContext(ctx), &request); err != nil {
		return c.wrapPgError(ctx, err, "failed to update funding schedule")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"fundingSchedule": request,
		"spending":        updatedSpending,
	})
}

func (c *Controller) deleteFundingSchedules(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fundingScheduleId, err := ParseID[FundingSchedule](ctx.Param("fundingScheduleId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid funding schedule Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	spending, err := repo.GetSpendingByFundingSchedule(c.getContext(ctx), bankAccountId, fundingScheduleId)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve spending for the funding schedule to be deleted")
	}

	if len(spending) > 0 {
		return c.badRequest(ctx, "Cannot delete a funding schedule with goals or expenses associated with it")
	}

	if err := repo.DeleteFundingSchedule(c.getContext(ctx), bankAccountId, fundingScheduleId); err != nil {
		if errors.Is(errors.Cause(err), repository.ErrFundingScheduleNotFound) {
			return c.notFound(ctx, "cannot remove funding schedule, it does not exist")
		}

		return c.wrapPgError(ctx, err, "failed to remove funding schedule")
	}

	return ctx.NoContent(http.StatusOK)
}
