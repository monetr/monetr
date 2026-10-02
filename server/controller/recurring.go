package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
	. "github.com/monetr/monetr/server/models"
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
