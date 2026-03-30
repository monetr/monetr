package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) getSimilarTransactionsByClusterId(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionClusterId, err := ParseID[TransactionCluster](ctx.Param("transactionClusterId"))
	if err != nil || transactionClusterId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction cluster Id")
	}

	limit := urlParamIntDefault(ctx, "limit", 10)
	offset := urlParamIntDefault(ctx, "offset", 0)

	if limit < 1 {
		return c.badRequest(ctx, "limit must be at least 1")
	} else if limit > 100 {
		return c.badRequest(ctx, "limit cannot be greater than 100")
	}

	if offset < 0 {
		return c.badRequest(ctx, "offset cannot be less than 0")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	transactions, err := repo.GetTransactionsByCluster(
		c.getContext(ctx),
		bankAccountId,
		transactionClusterId,
		limit,
		offset,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transactions")
	}

	return ctx.JSON(http.StatusOK, transactions)
}
