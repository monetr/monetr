package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/schemas"
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

	fundingSchedule := &FundingSchedule{
		BankAccountId: bankAccountId,
	}
	fundingSchedule, err = parse(
		c,
		ctx,
		fundingSchedule,
		schemas.CreateFundingSchedule,
	)
	if err != nil {
		return err
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// Make sure the bank account belongs to the authenticated user before
	// creating a funding schedule for it.
	if _, err = repo.GetBankAccount(c.getContext(ctx), bankAccountId); err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve bank account")
	}

	if fundingSchedule.AutoCreateTransaction {
		if fundingSchedule.EstimatedDeposit == nil || *fundingSchedule.EstimatedDeposit <= 0 {
			return c.badRequest(ctx, "Auto create transaction requires a non-zero estimated deposit")
		}
		isManual, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to validate if link is manual")
		}
		if !isManual {
			return c.badRequest(ctx, "Auto create transaction is only supported for manual links")
		}
	}

	// Set the next occurrence based on the provided rule. If the next occurrence
	// is not specified then assume that the rule is relative to now. If it is
	// specified though then do nothing, and the subsequent occurrence will be
	// calculated relative to the provided date. We also calculate the next
	// occurrence if the provided occurrence is in the past. This technically
	// should not happen via the UI. But it is currently possible for someone to
	// select the current day in the UI. Which then gets adjusted for midnight
	// that day, which will always be in the past for the user.
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

	if err = repo.CreateFundingSchedule(c.getContext(ctx), fundingSchedule); err != nil {
		return c.wrapPgError(ctx, err, "failed to create funding schedule")
	}

	return ctx.JSON(http.StatusOK, fundingSchedule)
}

func (c *Controller) patchFundingSchedule(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	fundingScheduleId, err := ParseID[FundingSchedule](ctx.Param("fundingScheduleId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid funding schedule Id")
	}

	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)
	// Retrieve the existing funding schedule to make sure some fields cannot be overridden
	originalFundingSchedule, err := repo.GetFundingSchedule(
		c.getContext(ctx),
		bankAccountId,
		fundingScheduleId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to verify funding schedule exists")
	}

	fundingSchedule, err := parse(
		c,
		ctx,
		originalFundingSchedule,
		schemas.PatchFundingSchedule,
	)
	if err != nil {
		return err
	}

	if fundingSchedule.NextRecurrence.Before(c.Clock.Now()) {
		return c.badRequest(ctx, "Next recurrence must be in the future")
	}

	if fundingSchedule.AutoCreateTransaction {
		if fundingSchedule.EstimatedDeposit == nil || *fundingSchedule.EstimatedDeposit <= 0 {
			return c.badRequest(ctx, "Auto create transaction requires a non-zero estimated deposit")
		}
		isManual, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to validate if link is manual")
		}
		if !isManual {
			return c.badRequest(ctx, "Auto create transaction is only supported for manual links")
		}
	}

	// Check if the changes made to the funding schedule would require that we
	// recalculate the contributions made to the spending objects that use the
	// funding schedule.
	var recalculateSpending bool
	if !originalFundingSchedule.NextRecurrence.Equal(fundingSchedule.NextRecurrence) {
		recalculateSpending = true
	}

	if originalFundingSchedule.RuleSet.String() != fundingSchedule.RuleSet.String() {
		recalculateSpending = true
	}

	if originalFundingSchedule.ExcludeWeekends != fundingSchedule.ExcludeWeekends {
		recalculateSpending = true
	}

	log = log.With("bankAccountId", bankAccountId, "fundingScheduleId", fundingScheduleId)

	updatedSpending := make([]Spending, 0)
	if recalculateSpending {
		log.DebugContext(c.getContext(ctx), "spending will be recalculated as part of this funding schedule update")
		crumbs.Debug(c.getContext(ctx), "Spending will be recalculated as part of this funding schedule update", map[string]any{
			"bankAccountId":     bankAccountId,
			"fundingScheduleId": fundingScheduleId,
		})
		spending, err := repo.GetSpendingByFundingSchedule(
			c.getContext(ctx),
			bankAccountId,
			fundingScheduleId,
		)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to retrieve existing spending objects for funding schedule")
		}

		if len(spending) > 0 {
			now := c.Clock.Now()
			timezone := c.mustGetTimezone(ctx)
			for _, spend := range spending {
				log.With("spendingId", spend.SpendingId).
					DebugContext(c.getContext(ctx), "recalculating next contribution for spending due to updated funding schedule")
				spend.CalculateNextContribution(
					c.getContext(ctx),
					timezone,
					fundingSchedule,
					now,
					log,
				)
			}

			if err = repo.UpdateSpending(
				c.getContext(ctx),
				bankAccountId,
				spending,
			); err != nil {
				return c.wrapPgError(ctx, err, "failed to update spending objects for updated funding schedule")
			}
			updatedSpending = spending
		}
	}

	if err = repo.UpdateFundingSchedule(c.getContext(ctx), fundingSchedule); err != nil {
		return c.wrapPgError(ctx, err, "failed to update funding schedule")
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"fundingSchedule": fundingSchedule,
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
