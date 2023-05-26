package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/models"
)

// List Spending
// @id list-spending
// @tags Spending
// @Summary List Spending
// @description List all of the spending for the specified bank account.
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Router /bank_accounts/{bankAccountId}/spending [get]
// @Success 200 {array} swag.SpendingResponse
// @Failure 400 {object} InvalidBankAccountIdError Invalid Bank Account ID.
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getSpending(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	expenses, err := repo.GetSpending(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retrieve expenses")
	}

	return ctx.JSON(http.StatusOK, expenses)
}

// Create Spending
// @id create-spending
// @tags Spending
// @Summary Create Spending
// @description Create an spending for the specified bank account.
// @security ApiKeyAuth
// @accept json
// @Produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Param Spending body swag.NewSpendingRequest true "New spending"
// @Router /bank_accounts/{bankAccountId}/spending [post]
// @Success 200 {object} swag.SpendingResponse
// @Failure 400 {object} InvalidBankAccountIdError "Invalid Bank Account ID."
// @Failure 400 {object} ApiError "Malformed JSON or invalid RRule."
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError "Failed to persist data."
func (c *Controller) postSpending(ctx echo.Context) error {
	requestSpan := c.getSpan(ctx)
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		requestSpan.Status = sentry.SpanStatusInvalidArgument
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	spending := &models.Spending{}
	if err := ctx.Bind(spending); err != nil {
		requestSpan.Status = sentry.SpanStatusInvalidArgument
		return c.invalidJson(ctx)
	}

	spending.SpendingId = 0 // Make sure we create a new spending.
	spending.BankAccountId = bankAccountId
	spending.Name = strings.TrimSpace(spending.Name)
	spending.Description = strings.TrimSpace(spending.Description)

	if spending.Name == "" {
		requestSpan.Status = sentry.SpanStatusInvalidArgument
		return c.badRequest(ctx, "spending must have a name")
	}

	if spending.TargetAmount <= 0 {
		requestSpan.Status = sentry.SpanStatusInvalidArgument
		return c.badRequest(ctx, "target amount must be greater than 0")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// We need to calculate what the next contribution will be for this new spending. So we need to retrieve it's funding
	// schedule. This also helps us validate that the user has provided a valid funding schedule id.
	fundingSchedule, err := repo.GetFundingSchedule(c.getContext(ctx), bankAccountId, spending.FundingScheduleId)
	if err != nil {
		requestSpan.Status = sentry.SpanStatusNotFound
		return c.wrapPgError(ctx, err, "could not find funding schedule specified")
	}

	// We also need to know the current account's timezone, as contributions are made at midnight in that user's
	// timezone.
	account, err := repo.GetAccount(c.getContext(ctx))
	if err != nil {
		requestSpan.Status = sentry.SpanStatusNotFound
		return c.wrapPgError(ctx, err, "failed to retrieve account details")
	}

	spending.LastRecurrence = nil

	var next time.Time

	switch spending.SpendingType {
	case models.SpendingTypeExpense:
		next = spending.NextRecurrence
		// Once we know that the next recurrence is not in the past we can just store it here;
		// itll be sanitized and converted to midnight below.
		if next.Before(time.Now()) {
			requestSpan.Status = sentry.SpanStatusInvalidArgument
			return c.badRequest(ctx, "next due date cannot be inthe past")
		}
	case models.SpendingTypeGoal:
		// If the spending is a goal, then we don't need the rule at all.
		next = spending.NextRecurrence
		if next.Before(time.Now()) {
			requestSpan.Status = sentry.SpanStatusInvalidArgument
			return c.badRequest(ctx, "due date cannot be in the past")
		}

		// Goals do not recur.
		spending.RecurrenceRule = nil
	}

	// Make sure that the next recurrence date is properly in the user's timezone.
	nextRecurrence, err := c.midnightInLocal(ctx, next)
	if err != nil {
		requestSpan.Status = sentry.SpanStatusInternalError
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not determine next recurrence")
	}

	spending.NextRecurrence = nextRecurrence

	// Once we have all that data we can calculate the new expenses next contribution amount.
	if err = spending.CalculateNextContribution(
		c.getContext(ctx),
		account.Timezone,
		fundingSchedule,
		time.Now(),
	); err != nil {
		requestSpan.Status = sentry.SpanStatusInternalError
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate the next contribution for the new spending")
	}

	if err = repo.CreateSpending(c.getContext(ctx), spending); err != nil {
		requestSpan.Status = sentry.SpanStatusInternalError
		return c.wrapPgError(ctx, err, "failed to create spending")
	}

	return ctx.JSON(http.StatusOK, spending)
}

type SpendingTransfer struct {
	FromSpendingId *uint64 `json:"fromSpendingId"`
	ToSpendingId   *uint64 `json:"toSpendingId"`
	Amount         int64   `json:"amount"`
}

// Transfer To or From Spending
// @id transfer-spending
// @tags Spending
// @Summary Transfer To or From Spending
// @description Transfer allocated funds to or from a spending object.
// @security ApiKeyAuth
// @accept json
// @produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Param Spending body SpendingTransfer true "Transfer"
// @Router /bank_accounts/{bankAccountId}/spending/transfer [post]
// @Success 200 {array} swag.TransferResponse
// @Failure 400 {object} InvalidBankAccountIdError "Invalid Bank Account ID."
// @Failure 400 {object} ApiError "Malformed JSON or invalid RRule."
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError "Failed to persist data."
func (c *Controller) postSpendingTransfer(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transfer := &SpendingTransfer{}
	if err := ctx.Bind(transfer); err != nil {
		return c.invalidJson(ctx)
	}

	if transfer.Amount <= 0 {
		return c.badRequest(ctx, "transfer amount must be greater than 0")
	}

	if (transfer.FromSpendingId == nil || *transfer.FromSpendingId == 0) &&
		(transfer.ToSpendingId == nil || *transfer.ToSpendingId == 0) {
		return c.badRequest(ctx, "both a from and a to must be specified to transfer allocated funds")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	balances, err := repo.GetBalances(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to get balances for transfer")
	}

	spendingToUpdate := make([]models.Spending, 0)

	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve account for transfer")
	}

	var fundingSchedule *models.FundingSchedule

	if transfer.FromSpendingId == nil && balances.Safe < transfer.Amount {
		return c.badRequest(ctx, "cannot transfer more than is available in safe to spend")
	} else if transfer.FromSpendingId != nil {
		fromExpense, err := repo.GetSpendingById(c.getContext(ctx), bankAccountId, *transfer.FromSpendingId)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to retrieve source expense for transfer")
		}

		if fromExpense.CurrentAmount < transfer.Amount {
			return c.badRequest(ctx, "cannot transfer more than is available in source goal/expense")
		}

		fundingSchedule, err = repo.GetFundingSchedule(c.getContext(ctx), bankAccountId, fromExpense.FundingScheduleId)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to retrieve funding schedule for source goal/expense")
		}

		fromExpense.CurrentAmount -= transfer.Amount

		if err = fromExpense.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			fundingSchedule,
			time.Now(),
		); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate next contribution for source goal/expense")
		}

		spendingToUpdate = append(spendingToUpdate, *fromExpense)
	}

	// If we are transferring the allocated funds to another spending object then we need to update that object. If we
	// are transferring it back to "Safe to spend" then we can just subtract the allocation from the source.
	if transfer.ToSpendingId != nil {
		toExpense, err := repo.GetSpendingById(c.getContext(ctx), bankAccountId, *transfer.ToSpendingId)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to get destination goal/expense for transfer")
		}

		// If the funding schedule that we already have put aside is not the same as the one we need for this spending
		// then we need to retrieve the proper one.
		if fundingSchedule == nil || fundingSchedule.FundingScheduleId != toExpense.FundingScheduleId {
			fundingSchedule, err = repo.GetFundingSchedule(c.getContext(ctx), bankAccountId, toExpense.FundingScheduleId)
			if err != nil {
				return c.wrapPgError(ctx, err, "failed to retrieve funding schedule for destination goal/expense")
			}
		}

		toExpense.CurrentAmount += transfer.Amount

		if err = toExpense.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			fundingSchedule,
			time.Now(),
		); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate next contribution for source goal/expense")
		}

		spendingToUpdate = append(spendingToUpdate, *toExpense)
	}

	if err = repo.UpdateSpending(c.getContext(ctx), bankAccountId, spendingToUpdate); err != nil {
		return c.wrapPgError(ctx, err, "failed to update spending for transfer")
	}

	balance, err := repo.GetBalances(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not get updated balances")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"balance":  balance,
		"spending": spendingToUpdate,
	})
}

// Update Spending
// @id update-spending
// @tags Spending
// @summary Update Spending
// @description Update an existing spending object. Some changes may cause the spending object to be recalculated.
// @security ApiKeyAuth
// @accept json
// @produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Param Spending body swag.UpdateSpendingRequest true "Updated spending"
// @Router /bank_accounts/{bankAccountId}/spending [put]
// @Success 200 {object} swag.SpendingResponse
// @Failure 400 {object} InvalidBankAccountIdError "Invalid Bank Account ID."
// @Failure 400 {object} ApiError "Malformed JSON or invalid RRule."
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError "Failed to persist data."
func (c *Controller) putSpending(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	spendingId, err := strconv.ParseUint(ctx.Param("spendingId"), 10, 64)
	if err != nil || spendingId == 0 {
		return c.badRequest(ctx, "must specify valid spending Id")
	}

	updatedSpending := &models.Spending{}
	if err := ctx.Bind(updatedSpending); err != nil {
		return c.invalidJson(ctx)
	}
	updatedSpending.SpendingId = spendingId
	updatedSpending.BankAccountId = bankAccountId

	repo := c.mustGetAuthenticatedRepository(ctx)

	existingSpending, err := repo.GetSpendingById(c.getContext(ctx), bankAccountId, updatedSpending.SpendingId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to find existing spending")
	}

	if updatedSpending.TargetAmount <= 0 {
		return c.badRequest(ctx, "target amount must be greater than 0")
	}

	// These fields cannot be changed by the end user and must be maintained by the API, some of these fields are
	// just meant to be immutable like date created.
	updatedSpending.SpendingType = existingSpending.SpendingType
	updatedSpending.DateCreated = existingSpending.DateCreated
	updatedSpending.UsedAmount = existingSpending.UsedAmount
	updatedSpending.CurrentAmount = existingSpending.CurrentAmount
	updatedSpending.BankAccountId = existingSpending.BankAccountId
	updatedSpending.IsBehind = existingSpending.IsBehind
	updatedSpending.LastRecurrence = existingSpending.LastRecurrence
	updatedSpending.NextContributionAmount = existingSpending.NextContributionAmount

	if updatedSpending.SpendingType == models.SpendingTypeGoal {
		updatedSpending.RecurrenceRule = nil
	}

	recalculateSpending := false
	if updatedSpending.NextRecurrence != existingSpending.NextRecurrence {
		newNext, err := c.midnightInLocal(ctx, updatedSpending.NextRecurrence)
		if err != nil {
			return c.badRequest(ctx, "failed to update next recurrence")
		}

		if newNext != existingSpending.NextRecurrence {
			updatedSpending.NextRecurrence = newNext
			recalculateSpending = true
		}
	}

	if updatedSpending.TargetAmount != existingSpending.TargetAmount {
		recalculateSpending = true
	} else if updatedSpending.FundingScheduleId != existingSpending.FundingScheduleId {
		recalculateSpending = true
	} else if !recalculateSpending && updatedSpending.RecurrenceRule != nil {
		recalculateSpending = updatedSpending.RecurrenceRule.String() == existingSpending.RecurrenceRule.String()
	}

	// If the paused status of a spending object changes, recalculate the contributions.
	if !updatedSpending.IsPaused && existingSpending.IsPaused {
		recalculateSpending = true
	} else if updatedSpending.IsPaused && !existingSpending.IsPaused {
		// However, if we are pausing contributions, there is no need to do a recalculation no matter what. Since it
		// will be invalidated when the user unpauses the spending object anyway.
		recalculateSpending = false
	}

	if recalculateSpending {
		account, err := repo.GetAccount(c.getContext(ctx))
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to retrieve account details")
		}

		fundingSchedule, err := repo.GetFundingSchedule(c.getContext(ctx), bankAccountId, updatedSpending.FundingScheduleId)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to retrieve funding schedule")
		}

		if err = updatedSpending.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			fundingSchedule,
			time.Now(),
		); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate next contribution")
		}
	}

	if err = repo.UpdateSpending(c.getContext(ctx), bankAccountId, []models.Spending{
		*updatedSpending,
	}); err != nil {
		return c.wrapPgError(ctx, err, "failed to update spending")
	}

	return ctx.JSON(http.StatusOK, updatedSpending)
}

// Delete Spending
// @id delete-spending
// @tags Spending
// @summary Delete Spending
// @description Delete a spending object. This will set any transactions that have spent from this object back to spent from "Safe-To-Spend". If the spending object is successfully deleted, this endpoint simply returns 200 with an empty body.
// @security ApiKeyAuth
// @accept json
// @produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Param spendingId path int true "Spending ID to be deleted"
// @Router /bank_accounts/{bankAccountId}/spending/{spendingId} [delete]
// @Success 200
// @Failure 400 {object} ApiError "Malformed JSON or invalid RRule."
// @Failure 500 {object} ApiError "Failed to persist data."
func (c *Controller) deleteSpending(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	spendingId, err := strconv.ParseUint(ctx.Param("spendingId"), 10, 64)
	if err != nil || spendingId == 0 {
		return c.badRequest(ctx, "must specify valid spending Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.DeleteSpending(c.getContext(ctx), bankAccountId, spendingId); err != nil {
		return c.wrapPgError(ctx, err, "failed to delete spending")
	}

	return ctx.NoContent(http.StatusOK)
}
