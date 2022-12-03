package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/monetr/monetr/pkg/models"
)

// @tag.name Spending
// @tag.description Spending endpoints handle the underlying spending object. The spending object is used to represent a goal or an expense.
func (c *Controller) handleSpending(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/spending", c.getSpending)
	p.Post("/{bankAccountId:uint64}/spending", c.postSpending)
	p.Post("/{bankAccountId:uint64}/spending/transfer", c.postSpendingTransfer)
	p.Put("/{bankAccountId:uint64}/spending/{spendingId:uint64}", c.putSpending)
	p.Delete("/{bankAccountId:uint64}/spending/{spendingId:uint64}", c.deleteSpending)
}

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
func (c *Controller) getSpending(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	expenses, err := repo.GetSpending(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not retrieve expenses")
		return
	}

	ctx.JSON(expenses)
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
func (c *Controller) postSpending(ctx *context.Context) {
	requestSpan := c.getSpan(ctx)
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		requestSpan.Status = sentry.SpanStatusInvalidArgument
		c.badRequest(ctx, "must specify valid bank account Id")
		return
	}

	var request struct {
		BankAccountId     uint64              `json:"bankAccountId"`
		FundingScheduleId uint64              `json:"fundingScheduleId"`
		SpendingType      models.SpendingType `json:"spendingType"`
		Name              string              `json:"name"`
		Description       string              `json:"description,omitempty"`
		TargetAmount      int64               `json:"targetAmount"`
		RecurrenceRule    *models.Rule        `json:"recurrenceRule"`
		NextRecurrence    time.Time           `json:"nextRecurrence"`
		IsPaused          bool                `json:"isPaused"`
	}
	if err := ctx.ReadJSON(request); err != nil {
		requestSpan.Status = sentry.SpanStatusInvalidArgument
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	spending := models.Spending{
		BankAccountId:   bankAccountId,
		SpendingFunding: []models.SpendingFunding{},
		SpendingType:    request.SpendingType,
		Name:            strings.TrimSpace(request.Name),
		Description:     strings.TrimSpace(request.Description),
		TargetAmount:    request.TargetAmount,
		CurrentAmount:   0,
		UsedAmount:      0,
		RecurrenceRule:  request.RecurrenceRule,
		LastRecurrence:  nil,
		NextRecurrence:  request.NextRecurrence,
		IsBehind:        false,
		IsPaused:        request.IsPaused,
		DateCreated:     time.Now().UTC(),
	}

	if spending.Name == "" {
		requestSpan.Status = sentry.SpanStatusInvalidArgument
		c.badRequest(ctx, "spending must have a name")
		return
	}

	if spending.TargetAmount <= 0 {
		requestSpan.Status = sentry.SpanStatusInvalidArgument
		c.badRequest(ctx, "target amount must be greater than 0")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// We need to calculate what the next contribution will be for this new spending. So we need to retrieve it's funding
	// schedule. This also helps us validate that the user has provided a valid funding schedule id.
	fundingSchedule, err := repo.GetFundingSchedule(
		c.getContext(ctx),
		bankAccountId,
		request.FundingScheduleId,
	)
	if err != nil {
		requestSpan.Status = sentry.SpanStatusNotFound
		c.wrapPgError(ctx, err, "could not find funding schedule specified")
		return
	}

	spendingFunding := models.SpendingFunding{
		BankAccountId:          bankAccountId,
		SpendingId:             0,
		FundingScheduleId:      request.FundingScheduleId,
		FundingSchedule:        fundingSchedule,
		NextContributionAmount: 0,
	}

	// We also need to know the current account's timezone, as contributions are made at midnight in that user's
	// timezone.
	account, err := repo.GetAccount(c.getContext(ctx))
	if err != nil {
		requestSpan.Status = sentry.SpanStatusNotFound
		c.wrapPgError(ctx, err, "failed to retrieve account details")
		return
	}

	spending.LastRecurrence = nil

	var next time.Time

	switch spending.SpendingType {
	case models.SpendingTypeExpense:
		// If this is an expense then we need to figure out when it happens next.
		next = spending.RecurrenceRule.After(time.Now(), false)
	case models.SpendingTypeGoal:
		// If the spending is a goal, then we don't need the rule at all.
		next = spending.NextRecurrence
		if next.Before(time.Now()) {
			requestSpan.Status = sentry.SpanStatusInvalidArgument
			c.badRequest(ctx, "due date cannot be in the past")
			return
		}

		// Goals do not recur.
		spending.RecurrenceRule = nil
	}

	// Make sure that the next recurrence date is properly in the user's timezone.
	nextRecurrence, err := c.midnightInLocal(ctx, next)
	if err != nil {
		requestSpan.Status = sentry.SpanStatusInternalError
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not determine next recurrence")
		return
	}

	spending.NextRecurrence = nextRecurrence

	// Once we have all that data we can calculate the new expenses next contribution amount.
	updatedFunding, err := spending.CalculateNextContribution(
		c.getContext(ctx),
		account.Timezone,
		models.NewSpendingFundingHelper([]models.SpendingFunding{
			spendingFunding,
		}),
		time.Now(),
	)
	if err != nil {
		requestSpan.Status = sentry.SpanStatusInternalError
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate the next contribution for the new spending")
		return
	}

	if err = repo.CreateSpending(c.getContext(ctx), &spending); err != nil {
		requestSpan.Status = sentry.SpanStatusInternalError
		c.wrapPgError(ctx, err, "failed to create spending")
		return
	}
	updatedFunding.SpendingId = spending.SpendingId
	if err = repo.CreateSpendingFunding(c.getContext(ctx), updatedFunding); err != nil {
		requestSpan.Status = sentry.SpanStatusInternalError
		c.wrapPgError(ctx, err, "failed to create spending funding")
		return
	}

	ctx.JSON(spending)
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
func (c *Controller) postSpendingTransfer(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "must specify a valid bank account Id")
		return
	}

	transfer := &SpendingTransfer{}
	if err := ctx.ReadJSON(transfer); err != nil {
		c.invalidJson(ctx)
		return
	}

	if transfer.Amount <= 0 {
		c.badRequest(ctx, "transfer amount must be greater than 0")
		return
	}

	if (transfer.FromSpendingId == nil || *transfer.FromSpendingId == 0) &&
		(transfer.ToSpendingId == nil || *transfer.ToSpendingId == 0) {
		c.badRequest(ctx, "both a from and a to must be specified to transfer allocated funds")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	balances, err := repo.GetBalances(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to get balances for transfer")
		return
	}

	spendingToUpdate := make([]models.Spending, 0)
	fundingToUpdate := make([]models.SpendingFunding, 0)

	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve account for transfer")
		return
	}

	if transfer.FromSpendingId == nil && balances.Safe < transfer.Amount {
		c.badRequest(ctx, "cannot transfer more than is available in safe to spend")
		return
	} else if transfer.FromSpendingId != nil {
		fromExpense, err := repo.GetSpendingById(c.getContext(ctx), bankAccountId, *transfer.FromSpendingId)
		if err != nil {
			c.wrapPgError(ctx, err, "failed to retrieve source expense for transfer")
			return
		}

		if fromExpense.CurrentAmount < transfer.Amount {
			c.badRequest(ctx, "cannot transfer more than is available in source goal/expense")
			return
		}

		fromExpense.CurrentAmount -= transfer.Amount

		updatedFunding, err := fromExpense.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			models.NewSpendingFundingHelper(fromExpense.SpendingFunding),
			time.Now(),
		)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate next contribution for source goal/expense")
			return
		}

		spendingToUpdate = append(spendingToUpdate, *fromExpense)
		if updatedFunding != nil {
			fundingToUpdate = append(fundingToUpdate, *updatedFunding)
		}
	}

	// If we are transferring the allocated funds to another spending object then we need to update that object. If we
	// are transferring it back to "Safe to spend" then we can just subtract the allocation from the source.
	if transfer.ToSpendingId != nil {
		toExpense, err := repo.GetSpendingById(c.getContext(ctx), bankAccountId, *transfer.ToSpendingId)
		if err != nil {
			c.wrapPgError(ctx, err, "failed to get destination goal/expense for transfer")
			return
		}

		toExpense.CurrentAmount += transfer.Amount

		updatedFunding, err := toExpense.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			models.NewSpendingFundingHelper(toExpense.SpendingFunding),
			time.Now(),
		)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate next contribution for source goal/expense")
			return
		}

		spendingToUpdate = append(spendingToUpdate, *toExpense)
		if updatedFunding != nil {
			fundingToUpdate = append(fundingToUpdate, *updatedFunding)
		}
	}

	if err = repo.UpdateSpending(c.getContext(ctx), bankAccountId, spendingToUpdate); err != nil {
		c.wrapPgError(ctx, err, "failed to update spending for transfer")
		return
	}

	if err = repo.UpdateSpendingFunding(c.getContext(ctx), bankAccountId, fundingToUpdate); err != nil {
		c.wrapPgError(ctx, err, "failed to update spending funding instructions for transfer")
		return
	}

	balance, err := repo.GetBalances(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not get updated balances")
		return
	}

	ctx.JSON(map[string]interface{}{
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
func (c *Controller) putSpending(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	spendingId := ctx.Params().GetUint64Default("spendingId", 0)
	if spendingId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid spending Id")
		return
	}

	updatedSpending := &models.Spending{}
	if err := ctx.ReadJSON(updatedSpending); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}
	updatedSpending.SpendingId = spendingId
	updatedSpending.BankAccountId = bankAccountId

	repo := c.mustGetAuthenticatedRepository(ctx)

	existingSpending, err := repo.GetSpendingById(c.getContext(ctx), bankAccountId, updatedSpending.SpendingId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to find existing spending")
		return
	}

	if updatedSpending.TargetAmount <= 0 {
		c.badRequest(ctx, "target amount must be greater than 0")
		return
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

	if updatedSpending.SpendingType == models.SpendingTypeGoal {
		updatedSpending.RecurrenceRule = nil
	}

	recalculateSpending := false
	if updatedSpending.NextRecurrence != existingSpending.NextRecurrence {
		newNext, err := c.midnightInLocal(ctx, updatedSpending.NextRecurrence)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to update next recurrence")
			return
		}

		if newNext != existingSpending.NextRecurrence {
			updatedSpending.NextRecurrence = newNext
			recalculateSpending = true
		}
	}

	if updatedSpending.TargetAmount != existingSpending.TargetAmount {
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
			c.wrapPgError(ctx, err, "failed to retrieve account details")
			return
		}

		updatedFunding, err := updatedSpending.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			models.NewSpendingFundingHelper(existingSpending.SpendingFunding),
			time.Now(),
		)
		if err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate next contribution")
			return
		}

		if updatedFunding != nil {
			if err = repo.UpdateSpendingFunding(c.getContext(ctx), bankAccountId, []models.SpendingFunding{
				*updatedFunding,
			}); err != nil {
				c.wrapPgError(ctx, err, "failed to update spending funding")
				return
			}
		}
	}

	if err = repo.UpdateSpending(c.getContext(ctx), bankAccountId, []models.Spending{
		*updatedSpending,
	}); err != nil {
		c.wrapPgError(ctx, err, "failed to update spending")
		return
	}

	ctx.JSON(updatedSpending)
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
func (c *Controller) deleteSpending(ctx iris.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid bank account Id")
		return
	}

	spendingId := ctx.Params().GetUint64Default("spendingId", 0)
	if spendingId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid spending Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.DeleteSpending(c.getContext(ctx), bankAccountId, spendingId); err != nil {
		c.wrapPgError(ctx, err, "failed to delete spending")
		return
	}

	return
}
