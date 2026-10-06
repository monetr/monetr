package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/schemas"
)

func (c *Controller) getRecurringTransaction(ctx *echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionRecurringId, err := ParseID[TransactionRecurring](ctx.Param("transactionRecurringId"))
	if err != nil || transactionRecurringId.IsZero() {
		return c.badRequest(ctx, "must specify a valid recurring transaction Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	result, err := repo.GetTransactionRecurringById(
		c.getContext(ctx),
		bankAccountId,
		transactionRecurringId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve recurring transaction")
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *Controller) patchRecurringTransaction(ctx *echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionRecurringId, err := ParseID[TransactionRecurring](ctx.Param("transactionRecurringId"))
	if err != nil || transactionRecurringId.IsZero() {
		return c.badRequest(ctx, "must specify a valid recurring transaction Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	existing, err := repo.GetTransactionRecurringById(
		c.getContext(ctx),
		bankAccountId,
		transactionRecurringId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve recurring transaction")
	}

	recurring, err := parse(
		c,
		ctx,
		existing,
		schemas.PatchTransactionRecurring,
	)
	if err != nil {
		return err
	}

	// Spending only tracks money leaving the account, and for now only expenses
	// can be linked. Goals will get their own treatment later.
	if recurring.SpendingId != nil {
		if recurring.Direction != DebitDirection {
			return c.badRequest(ctx, "spending can only be linked to a debit recurring transaction")
		}

		spending, err := repo.GetSpendingById(
			c.getContext(ctx),
			bankAccountId,
			*recurring.SpendingId,
		)
		if err != nil {
			return c.wrapPgError(ctx, err, "could not find spending specified")
		}

		if spending.SpendingType != SpendingTypeExpense {
			return c.badRequest(ctx, "only expenses can be linked to a recurring transaction")
		}
	}

	// Funding schedules only track money coming into the account.
	if recurring.FundingScheduleId != nil {
		if recurring.Direction != CreditDirection {
			return c.badRequest(ctx, "funding schedules can only be linked to a credit recurring transaction")
		}

		if _, err := repo.GetFundingSchedule(
			c.getContext(ctx),
			bankAccountId,
			*recurring.FundingScheduleId,
		); err != nil {
			return c.wrapPgError(ctx, err, "could not find funding schedule specified")
		}
	}

	if err := repo.UpdateTransactionRecurring(
		c.getContext(ctx),
		bankAccountId,
		recurring,
	); err != nil {
		return c.wrapPgError(ctx, err, "failed to update recurring transaction")
	}

	// Read it back so the embedded spending and funding schedule match the new
	// links.
	result, err := repo.GetTransactionRecurringById(
		c.getContext(ctx),
		bankAccountId,
		transactionRecurringId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve updated recurring transaction")
	}

	return ctx.JSON(http.StatusOK, result)
}
