package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) getSpending(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
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

// getSpendingById serves a spending object by its specific ID, eventually it
// will also support serving soft-deleted spending items that might not be
// present in the index endpoint for spending.
func (c *Controller) getSpendingById(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	spendingId, err := ParseID[Spending](ctx.Param("spendingId"))
	if err != nil || spendingId.IsZero() {
		return c.badRequest(ctx, "must specify a valid spending Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	spending, err := repo.GetSpendingById(c.getContext(ctx), bankAccountId, spendingId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not retrieve spending")
	}
	// Unset this for the API.
	spending.FundingSchedule = nil

	return ctx.JSON(http.StatusOK, spending)
}

func (c *Controller) postSpending(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	spending := &Spending{}
	if err := ctx.Bind(spending); err != nil {
		return c.invalidJson(ctx)
	}

	spending.SpendingId = "" // Make sure we create a new spending.
	spending.BankAccountId = bankAccountId
	spending.Name, err = c.cleanString(ctx, "Name", spending.Name)
	if err != nil {
		return err
	}
	spending.Description, err = c.cleanString(ctx, "Description", spending.Description)
	if err != nil {
		return err
	}
	if spending.Name == "" {
		return c.badRequest(ctx, "spending must have a name")
	}

	if spending.TargetAmount <= 0 {
		return c.badRequest(ctx, "target amount must be greater than 0")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	// We need to calculate what the next contribution will be for this new
	// spending. So we need to retrieve it's funding schedule. This also helps us
	// validate that the user has provided a valid funding schedule id.
	fundingSchedule, err := repo.GetFundingSchedule(
		c.getContext(ctx),
		bankAccountId,
		spending.FundingScheduleId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not find funding schedule specified")
	}

	// We also need to know the current account's timezone, as contributions are
	// made at midnight in that user's timezone.
	account, err := repo.GetAccount(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve account details")
	}

	spending.LastRecurrence = nil

	// Once we know that the next recurrence is not in the past we can just store
	// it here; itll be sanitized and converted to midnight below.
	next := spending.NextRecurrence
	if next.Before(c.Clock.Now()) {
		return c.badRequest(ctx, "next due date cannot be in the past")
	}

	switch spending.SpendingType {
	case SpendingTypeExpense:
		if spending.RuleSet == nil {
			return c.badRequest(ctx, "recurrence rule must be specified for expenses")
		}
	case SpendingTypeGoal:
		if spending.RuleSet != nil {
			return c.badRequest(ctx, "recurrence rule cannot be specified for goals")
		}
	}

	// Make sure that the next recurrence date is properly in the user's timezone.
	nextRecurrence, err := c.midnightInLocal(ctx, next)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not determine next recurrence")
	}

	spending.NextRecurrence = nextRecurrence

	// Once we have all that data we can calculate the new expenses next contribution amount.
	if err = spending.CalculateNextContribution(
		c.getContext(ctx),
		account.Timezone,
		fundingSchedule,
		c.Clock.Now(),
	); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate the next contribution for the new spending")
	}

	if err = repo.CreateSpending(c.getContext(ctx), spending); err != nil {
		return c.wrapPgError(ctx, err, "failed to create spending")
	}

	return ctx.JSON(http.StatusOK, spending)
}

func (c *Controller) postSpendingTransfer(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	var transfer struct {
		FromSpendingId *ID[Spending] `json:"fromSpendingId"`
		ToSpendingId   *ID[Spending] `json:"toSpendingId"`
		Amount         int64         `json:"amount"`
	}
	if err := ctx.Bind(&transfer); err != nil {
		return c.invalidJson(ctx)
	}

	if transfer.Amount <= 0 {
		return c.badRequest(ctx, "transfer amount must be greater than 0")
	}

	if (transfer.FromSpendingId == nil || (*transfer.FromSpendingId).IsZero()) &&
		(transfer.ToSpendingId == nil || (*transfer.ToSpendingId).IsZero()) {
		return c.badRequest(ctx, "Both a from and a to must be specified to transfer allocated funds")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	spendingToUpdate := make([]Spending, 0, 2)

	account, err := c.Accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve account for transfer")
	}

	var fundingSchedule *FundingSchedule

	if transfer.FromSpendingId != nil {
		fromExpense, err := repo.GetSpendingById(c.getContext(ctx), bankAccountId, *transfer.FromSpendingId)
		if err != nil {
			return c.wrapPgError(ctx, err, "Failed to retrieve source expense for transfer")
		}

		if fromExpense.CurrentAmount < transfer.Amount {
			return c.badRequest(ctx, "Cannot transfer more than is available in source goal/expense")
		}

		fundingSchedule, err = repo.GetFundingSchedule(c.getContext(ctx), bankAccountId, fromExpense.FundingScheduleId)
		if err != nil {
			return c.wrapPgError(ctx, err, "Failed to retrieve funding schedule for source goal/expense")
		}

		fromExpense.CurrentAmount -= transfer.Amount

		if err = fromExpense.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			fundingSchedule,
			c.Clock.Now(),
		); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate next contribution for source goal/expense")
		}

		spendingToUpdate = append(spendingToUpdate, *fromExpense)
	}

	// If we are transferring the allocated funds to another spending object then
	// we need to update that object. If we are transferring it back to "Safe to
	// spend" then we can just subtract the allocation from the source.
	if transfer.ToSpendingId != nil {
		toExpense, err := repo.GetSpendingById(
			c.getContext(ctx),
			bankAccountId,
			*transfer.ToSpendingId,
		)
		if err != nil {
			return c.wrapPgError(ctx, err, "Failed to get destination goal/expense for transfer")
		}

		// If the funding schedule that we already have put aside is not the same as
		// the one we need for this spending then we need to retrieve the proper
		// one.
		if fundingSchedule == nil || fundingSchedule.FundingScheduleId != toExpense.FundingScheduleId {
			fundingSchedule, err = repo.GetFundingSchedule(
				c.getContext(ctx),
				bankAccountId,
				toExpense.FundingScheduleId,
			)
			if err != nil {
				return c.wrapPgError(ctx, err, "Failed to retrieve funding schedule for destination goal/expense")
			}
		}

		toExpense.CurrentAmount += transfer.Amount

		if err = toExpense.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			fundingSchedule,
			c.Clock.Now(),
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

func (c *Controller) putSpending(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	spendingId, err := ParseID[Spending](ctx.Param("spendingId"))
	if err != nil || spendingId.IsZero() {
		return c.badRequest(ctx, "must specify a valid spending Id")
	}

	updatedSpending := &Spending{}
	if err := ctx.Bind(updatedSpending); err != nil {
		return c.invalidJson(ctx)
	}
	updatedSpending.SpendingId = spendingId
	updatedSpending.BankAccountId = bankAccountId
	updatedSpending.Name, err = c.cleanString(ctx, "Name", updatedSpending.Name)
	if err != nil {
		return err
	}
	updatedSpending.Description, err = c.cleanString(ctx, "Description", updatedSpending.Description)
	if err != nil {
		return err
	}

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
	updatedSpending.CreatedAt = existingSpending.CreatedAt
	updatedSpending.UsedAmount = existingSpending.UsedAmount
	updatedSpending.CurrentAmount = existingSpending.CurrentAmount
	updatedSpending.BankAccountId = existingSpending.BankAccountId
	updatedSpending.IsBehind = existingSpending.IsBehind
	updatedSpending.LastRecurrence = existingSpending.LastRecurrence
	updatedSpending.NextContributionAmount = existingSpending.NextContributionAmount

	if updatedSpending.SpendingType == SpendingTypeGoal {
		updatedSpending.RuleSet = nil
	}

	if updatedSpending.SpendingType == SpendingTypeExpense && updatedSpending.RuleSet == nil {
		return c.badRequest(ctx, "Expense must have a recurrence rule provided")
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
	} else if !recalculateSpending && updatedSpending.RuleSet != nil {
		recalculateSpending = updatedSpending.RuleSet.String() == existingSpending.RuleSet.String()
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

		fundingSchedule, err := repo.GetFundingSchedule(
			c.getContext(ctx),
			bankAccountId,
			updatedSpending.FundingScheduleId,
		)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to retrieve funding schedule")
		}

		if err = updatedSpending.CalculateNextContribution(
			c.getContext(ctx),
			account.Timezone,
			fundingSchedule,
			c.Clock.Now(),
		); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to calculate next contribution")
		}
	}

	if err = repo.UpdateSpending(c.getContext(ctx), bankAccountId, []Spending{
		*updatedSpending,
	}); err != nil {
		return c.wrapPgError(ctx, err, "failed to update spending")
	}

	return ctx.JSON(http.StatusOK, updatedSpending)
}

func (c *Controller) deleteSpending(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	spendingId, err := ParseID[Spending](ctx.Param("spendingId"))
	if err != nil || spendingId.IsZero() {
		return c.badRequest(ctx, "must specify a valid spending Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	if err := repo.DeleteSpending(c.getContext(ctx), bankAccountId, spendingId); err != nil {
		return c.wrapPgError(ctx, err, "failed to delete spending")
	}

	return ctx.NoContent(http.StatusOK)
}

func (c *Controller) getSpendingTransactions(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	spendingId, err := ParseID[Spending](ctx.Param("spendingId"))
	if err != nil || spendingId.IsZero() {
		return c.badRequest(ctx, "must specify a valid spending Id")
	}

	limit := urlParamIntDefault(ctx, "limit", 25)
	offset := urlParamIntDefault(ctx, "offset", 0)

	if limit < 1 {
		return c.badRequest(ctx, "limit must be at least 1")
	} else if limit > 100 {
		return c.badRequest(ctx, "limit cannot be greater than 100")
	}

	if offset < 0 {
		return c.badRequest(ctx, "offset cannot be less than 0")
	}

	// Only let a maximum of 100 transactions be requested at a time.
	limit = myownsanity.Min(100, limit)

	repo := c.mustGetAuthenticatedRepository(ctx)

	ok, err := repo.GetSpendingExists(c.getContext(ctx), bankAccountId, spendingId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to verify spending exists")
	}

	if !ok {
		return c.returnError(ctx, http.StatusNotFound, "spending object does not exist")
	}

	transactions, err := repo.GetTransactionsForSpending(
		c.getContext(ctx),
		bankAccountId,
		spendingId,
		limit,
		offset,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transactions for spending")
	}

	return ctx.JSON(http.StatusOK, transactions)
}
