package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
)

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
func (c *Controller) getFundingSchedules(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil || bankAccountId == 0 {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	fundingSchedules, err := repo.GetFundingSchedules(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve funding schedules")
	}

	if fundingSchedules == nil {
		fundingSchedules = make([]models.FundingSchedule, 0)
	}

	return ctx.JSON(http.StatusOK, fundingSchedules)
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
func (c *Controller) getFundingScheduleStats(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil || bankAccountId == 0 {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	stats, err := repo.GetFundingStats(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve funding schedules")
	}

	return ctx.JSON(http.StatusOK, stats)
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
func (c *Controller) postFundingSchedules(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil || bankAccountId == 0 {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	var fundingSchedule models.FundingSchedule
	if err := ctx.Bind(&fundingSchedule); err != nil {
		return c.invalidJson(ctx)
	}

	fundingSchedule.FundingScheduleId = 0 // Make sure we create a new funding schedule.
	fundingSchedule.Name = strings.TrimSpace(fundingSchedule.Name)
	fundingSchedule.Description = strings.TrimSpace(fundingSchedule.Description)
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
	if (time.Time{}).Equal(fundingSchedule.NextOccurrence) || time.Now().After(fundingSchedule.NextOccurrence) {
		fundingSchedule.CalculateNextOccurrence(c.getContext(ctx), c.mustGetTimezone(ctx))
	}

	// It has never occurred so this needs to be nil.
	fundingSchedule.LastOccurrence = nil
	fundingSchedule.DateStarted = fundingSchedule.NextOccurrence

	if err = repo.CreateFundingSchedule(c.getContext(ctx), &fundingSchedule); err != nil {
		return c.wrapPgError(ctx, err, "failed to create funding schedule")
	}

	return ctx.JSON(http.StatusOK, fundingSchedule)
}

func (c *Controller) putFundingSchedules(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil || bankAccountId == 0 {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fundingScheduleId, err := strconv.ParseUint(ctx.Param("fundingScheduleId"), 10, 64)
	if err != nil || fundingScheduleId == 0 {
		return c.badRequest(ctx, "must specify valid funding schedule Id")
	}
	var request models.FundingSchedule
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	// Retrieve the existing funding schedule to make sure some fields cannot be overridden
	existingFundingSchedule, err := repo.GetFundingSchedule(c.getContext(ctx), bankAccountId, fundingScheduleId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to verify funding schedule exists")
	}

	request.FundingScheduleId = fundingScheduleId
	request.BankAccountId = bankAccountId
	request.Name = strings.TrimSpace(request.Name)
	request.Description = strings.TrimSpace(request.Description)
	request.AccountId = existingFundingSchedule.AccountId
	request.LastOccurrence = existingFundingSchedule.LastOccurrence

	if request.Name == "" {
		return c.badRequest(ctx, "funding schedule must have a name")
	}

	if request.Rule == nil {
		return c.badRequest(ctx, "funding schedule must include a rule")
	}

	if request.EstimatedDeposit != nil && *request.EstimatedDeposit < 0 {
		return c.badRequest(ctx, "estimated deposit must be greater than or equal to zero")
	}

	recalculateSpending := false
	// If the next occurrence changes then we need to recalulate spending.
	if !request.NextOccurrence.Equal(existingFundingSchedule.NextOccurrence) {
		// The user cannot override the next occurrence for a funding schedule and have it be in the past. If they set it to
		// be in the future then that is okay. The next time the funding schedule is processed it will be relative to that
		// next occurrence.
		if request.NextOccurrence.Before(time.Now()) {
			request.NextOccurrence = existingFundingSchedule.NextOccurrence
		}
		recalculateSpending = true
	}

	// If the recurrence rule has changed then we need to recalculate spending too.
	if request.Rule.String() != existingFundingSchedule.Rule.String() {
		recalculateSpending = true
	}

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
		}
	}

	if err = repo.UpdateFundingSchedule(c.getContext(ctx), &request); err != nil {
		return c.wrapPgError(ctx, err, "failed to update funding schedule")
	}

	return ctx.JSON(http.StatusOK, request)
}

func (c *Controller) deleteFundingSchedules(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fundingScheduleId, err := strconv.ParseUint(ctx.Param("fundingScheduleId"), 10, 64)
	if err != nil || fundingScheduleId == 0 {
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
