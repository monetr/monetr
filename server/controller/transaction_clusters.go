package controller

import (
	"net/http"

	"github.com/labstack/echo/v5"
	. "github.com/monetr/monetr/server/models"
)

func (c *Controller) getSimilarTransactions(ctx *echo.Context) error {
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

	repo := c.mustGetAuthenticatedRepository(ctx)

	items, err := repo.GetTransactionClusters(
		c.getContext(ctx),
		bankAccountId,
		limit,
		offset,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "Failed to retrieve similar transactions")
	}

	return ctx.JSON(http.StatusOK, items)
}

func (c *Controller) getSimilarTransactionsByClusterId(ctx *echo.Context) error {
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

func (c *Controller) getSimilarTransactionCluster(ctx *echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionClusterId, err := ParseID[TransactionCluster](ctx.Param("transactionClusterId"))
	if err != nil || transactionClusterId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction cluster Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	transactionCluster, err := repo.GetTransactionCluster(
		c.getContext(ctx),
		bankAccountId,
		transactionClusterId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transaction cluster")
	}

	return ctx.JSON(http.StatusOK, transactionCluster)
}
