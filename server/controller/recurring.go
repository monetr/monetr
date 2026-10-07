package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/schemas"
)

func (c *Controller) getRecurringTransactions(ctx *echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "Must specify a valid bank account Id")
	}

	limit := urlParamIntDefault(ctx, "limit", 25)
	offset := urlParamIntDefault(ctx, "offset", 0)

	if limit < 1 {
		return c.badRequest(ctx, "Limit must be at least 1")
	} else if limit > 100 {
		return c.badRequest(ctx, "Limit cannot be greater than 100")
	}

	if offset < 0 {
		return c.badRequest(ctx, "Offset cannot be less than 0")
	}

	// These are both optional, leaving them off just doesn't filter on them. This
	// way the UI can ask for only the tab it is showing instead of everything.
	var direction *Direction
	if raw := ctx.QueryParam("direction"); raw != "" {
		switch Direction(raw) {
		case DebitDirection, CreditDirection:
			direction = new(Direction(raw))
		default:
			return c.badRequest(ctx, "Direction must be debit or credit")
		}
	}

	var ended *bool
	if raw := ctx.QueryParam("ended"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return c.badRequest(ctx, "Ended must be true or false")
		}
		ended = new(value)
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	items, err := repo.GetTransactionRecurrings(
		c.getContext(ctx),
		bankAccountId,
		direction,
		ended,
		limit,
		offset,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve recurring transactions")
	}

	return ctx.JSON(http.StatusOK, items)
}

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

	// Whatever the links are now, the user chose them.
	recurring.AutoMatched = false

	if err := repo.UpdateTransactionRecurring(
		c.getContext(ctx),
		bankAccountId,
		recurring,
	); err != nil {
		return c.wrapPgError(ctx, err, "failed to update recurring transaction")
	}

	// Read it back so the embedded funding schedule matches the new
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
