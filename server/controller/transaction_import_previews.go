package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) getTransactionImportPreviewByTransactionImportId(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionImportId, err := ParseID[TransactionImport](ctx.Param("transactionImportId"))
	if err != nil || transactionImportId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction import Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	existing, err := repo.GetTransactionImportPreviewByTransactionImportId(
		c.getContext(ctx),
		bankAccountId,
		transactionImportId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transaction import preview")
	}

	return ctx.JSON(http.StatusOK, existing)
}
