package controller

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"math"
	"net/http"
)

func (c *Controller) handleTransactions(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/transactions", c.getTransactions)
	p.Post("/{bankAccountId:uint64}/transactions", c.postTransactions)
	p.Put("/{bankAccountId:uint64}/transactions/{transactionId:uint64}", c.putTransactions)
	p.Delete("/{bankAccountId:uint64}/transactions/{transactionId:uint64}", c.deleteTransactions)
}

func (c *Controller) getTransactions(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "must specify a valid bank account Id")
		return
	}

	limit := ctx.URLParamIntDefault("limit", 25)
	offset := ctx.URLParamIntDefault("offset", 0)

	// Only let a maximum of 100 transactions be requested at a time.
	limit = int(math.Min(100, float64(limit)))

	repo := c.mustGetAuthenticatedRepository(ctx)

	transactions, err := repo.GetTransactions(bankAccountId, limit, offset)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve transactions")
		return
	}

	ctx.JSON(map[string]interface{}{
		"transactions": transactions,
	})
}

func (c *Controller) postTransactions(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "must specify a valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	isManual, err := repo.GetLinkIsManualByBankAccountId(bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to validate if link is manual")
		return
	}

	if !isManual {
		c.returnError(ctx, http.StatusBadRequest, "cannot create transactions for non-manual links")
		return
	}

}

func (c *Controller) putTransactions(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "must specify a valid bank account Id")
		return
	}

	transactionId := ctx.Params().GetUint64Default("transactionId", 0)
	if transactionId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid transaction Id")
		return
	}

	var transaction models.Transaction
	if err := ctx.ReadJSON(&transaction); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	transaction.TransactionId = transactionId
	transaction.BankAccountId = bankAccountId

	repo := c.mustGetAuthenticatedRepository(ctx)

	isManual, err := repo.GetLinkIsManualByBankAccountId(bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to validate if link is manual")
		return
	}

	existingTransaction, err := repo.GetTransaction(bankAccountId, transactionId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve existing transaction for update")
		return
	}

	if !isManual {
		// Prevent the user from attempting to change a transaction's amount if we are on a plaid link.
		if existingTransaction.Amount != transaction.Amount {
			c.returnError(ctx, http.StatusBadRequest, "cannot change transaction amount on non-manual links")
			return
		}

		if existingTransaction.IsPending != transaction.IsPending {
			c.badRequest(ctx, "cannot change transaction pending state on non-manual links")
			return
		}

		if existingTransaction.Date != transaction.Date {
			c.badRequest(ctx, "cannot change transaction date on non-manual links")
			return
		}

		if existingTransaction.AuthorizedDate != transaction.AuthorizedDate {
			c.badRequest(ctx, "cannot change transaction authorized date on non-manual links")
			return
		}
	}

}

func (c *Controller) deleteTransactions(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "must specify a valid bank account Id")
		return
	}

	transactionId := ctx.Params().GetUint64Default("transactionId", 0)
	if transactionId == 0 {
		c.returnError(ctx, http.StatusBadRequest, "must specify valid transaction Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	isManual, err := repo.GetLinkIsManualByBankAccountId(bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to validate if link is manual")
		return
	}

	if !isManual {
		c.returnError(ctx, http.StatusBadRequest, "cannot delete transactions for non-manual links")
		return
	}
}
