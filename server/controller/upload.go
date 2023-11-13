package controller

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func (c *Controller) postUploadTransactions(ctx echo.Context) error {
	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	ok, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to verify bank account link type")
	}

	if !ok {
		return c.badRequest(ctx, "Cannot import transactions for non-manual link.")
	}

	return nil
}
