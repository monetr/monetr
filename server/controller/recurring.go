package controller

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/recurring"
	"github.com/sirupsen/logrus"
)

func (c *Controller) getRecurring(ctx echo.Context) error {
	log := c.getLog(ctx)

	bankAccountId, err := strconv.ParseUint(ctx.Param("bankAccountId"), 10, 64)
	if err != nil {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	timezone := c.mustGetTimezone(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)

	recurringDetection := recurring.NewRecurringTransactionDetection(timezone)

	limit := 100
	offset := 0
	for {
		txnLog := log.WithFields(logrus.Fields{
			"limit":  limit,
			"offset": offset,
		})
		txnLog.Trace("requesting next batch of transactions")
		transactions, err := repo.GetTransactions(c.getContext(ctx), bankAccountId, limit, offset)
		if err != nil {
			return c.wrapPgError(ctx, err, "failed to read transactions")
		}
		txnLog = log.WithField("count", len(transactions))

		for i := range transactions {
			recurringDetection.AddTransaction(&transactions[i])
		}

		if len(transactions) < limit {
			txnLog.Trace("reached end of transactions")
			break
		}

		offset += len(transactions)
	}

	result := recurringDetection.GetRecurringTransactions()

	return ctx.JSON(http.StatusOK, result)
}
