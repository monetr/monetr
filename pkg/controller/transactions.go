package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/sirupsen/logrus"
	"math"
	"net/http"
	"strings"
)

func (c *Controller) handleTransactions(p iris.Party) {
	p.Get("/{bankAccountId:uint64}/transactions", c.getTransactions)
	p.Get("/{bankAccountId:uint64/transactions/spending/{spendingId:uint64}", c.getTransactionsForSpending)
	p.Post("/{bankAccountId:uint64}/transactions", c.postTransactions)
	p.Put("/{bankAccountId:uint64}/transactions/{transactionId:uint64}", c.putTransactions)
	p.Delete("/{bankAccountId:uint64}/transactions/{transactionId:uint64}", c.deleteTransactions)
}

// List Transactions
// @Summary List Transactions
// @ID list-transactions
// @tags Transactions
// @description Lists the transactions for the specified bank account Id. Transactions are returned sorted by the date
// @description they were authorized (descending) and then by their numeric Id (descending). This means that
// @description transactions that were first seen later will be higher in the list than they may have actually occurred.
// @Security ApiKeyAuth
// @Produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Param limit query int false "Specifies the number of transactions to return in the result, default is 25. Max is 100."
// @Param offset query int false "The number of transactions to skip before returning any."
// @Router /bank_accounts/{bankAccountId}/transactions [get]
// @Success 200 {array} swag.TransactionResponse
// @Failure 400 {object} InvalidBankAccountIdError Invalid Bank Account ID.
// @Failure 500 {object} ApiError Something went wrong on our end.
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

	transactions, err := repo.GetTransactions(c.getContext(ctx), bankAccountId, limit, offset)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve transactions")
		return
	}

	// If transactions are null or empty then make sure what we return is an empty array. Otherwise we can accidentally
	// return null.
	if len(transactions) == 0 {
		ctx.JSON(make([]models.Transaction, 0))
		return
	}

	ctx.JSON(transactions)
}

func (c *Controller) getTransactionsForSpending(ctx *context.Context) {

}

func (c *Controller) postTransactions(ctx *context.Context) {
	bankAccountId := ctx.Params().GetUint64Default("bankAccountId", 0)
	if bankAccountId == 0 {
		c.badRequest(ctx, "must specify a valid bank account Id")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	isManual, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to validate if link is manual")
		return
	}

	if !isManual {
		c.returnError(ctx, http.StatusBadRequest, "cannot create transactions for non-manual links")
		return
	}

	var transaction models.Transaction
	if err = ctx.ReadJSON(&transaction); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	transaction.BankAccountId = bankAccountId
	transaction.Name = strings.TrimSpace(transaction.Name)
	transaction.MerchantName = strings.TrimSpace(transaction.MerchantName)
	transaction.OriginalName = transaction.Name

	if transaction.Name == "" {
		c.badRequest(ctx, "transaction must have a name")
		return
	}

	if transaction.Amount <= 0 {
		c.badRequest(ctx, "transaction amount must be greater than 0")
		return
	}

	var updatedExpense *models.Spending

	if transaction.SpendingId != nil && *transaction.SpendingId > 0 {
		updatedExpense, err = repo.GetSpendingById(c.getContext(ctx), bankAccountId, *transaction.SpendingId)
		if err != nil {
			c.wrapPgError(ctx, err, "could not get expense provided for transaction")
			return
		}

		if err = repo.AddExpenseToTransaction(c.getContext(ctx), &transaction, updatedExpense); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to add expense to transaction")
			return
		}

		if err = repo.UpdateSpending(c.getContext(ctx), bankAccountId, []models.Spending{
			*updatedExpense,
		}); err != nil {
			c.wrapPgError(ctx, err, "failed to update expense for transaction")
			return
		}
	}

	if err = repo.CreateTransaction(c.getContext(ctx), bankAccountId, &transaction); err != nil {
		c.wrapPgError(ctx, err, "could not create transaction")
		return
	}

	returnedObject := map[string]interface{}{
		"transaction": transaction,
	}

	// If an expense was updated as part of this transaction being created then we want to include that updated expense
	// in our response so the UI can update its redux store.
	if updatedExpense != nil {
		returnedObject["expense"] = *updatedExpense
	}

	ctx.JSON(returnedObject)
}

// Update Transaction
// @Summary Update Transaction
// @ID update-transactions
// @tags Transactions
// @description Updates the provided transaction.
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param bankAccountId path int true "Bank Account ID"
// @Param transactionId path int true "TransactionId "
// @Param Transaction body swag.UpdateTransactionRequest true "Updated transaction"
// @Router /bank_accounts/{bankAccountId}/transactions/{transactionId} [post]
// @Success 200 {array} swag.TransactionResponse
// @Failure 400 {object} InvalidBankAccountIdError Invalid Bank Account ID.
// @Failure 500 {object} ApiError Something went wrong on our end.
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

	isManual, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to validate if link is manual")
		return
	}

	existingTransaction, err := repo.GetTransaction(c.getContext(ctx), bankAccountId, transactionId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve existing transaction for update")
		return
	}

	transaction.PlaidTransactionId = existingTransaction.PlaidTransactionId

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

		if !existingTransaction.Date.Equal(transaction.Date) {
			c.getLog(ctx).WithFields(logrus.Fields{
				"existingDate": existingTransaction.Date,
				"newDate":      transaction.Date,
			}).Warn("cannot change transaction date on non-manual links")
			c.badRequest(ctx, "cannot change transaction date on non-manual links")
			return
		}

		if !myownsanity.TimesPEqual(existingTransaction.AuthorizedDate, transaction.AuthorizedDate) {
			c.getLog(ctx).WithFields(logrus.Fields{
				"existingAuthorizedDate": existingTransaction.AuthorizedDate,
				"newAuthorizedDate":      transaction.AuthorizedDate,
			}).Warn("cannot change transaction authorized date on non-manual links")
			c.badRequest(ctx, "cannot change transaction authorized date on non-manual links")
			return
		}

		transaction.OriginalName = existingTransaction.OriginalName
		transaction.OriginalMerchantName = existingTransaction.OriginalMerchantName
		transaction.OriginalCategories = existingTransaction.OriginalCategories
	}

	updatedExpenses, err := repo.ProcessTransactionSpentFrom(c.getContext(ctx), bankAccountId, &transaction, existingTransaction)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to process expense changes")
		return
	}

	// TODO Handle more complex transaction updates via the API.
	//  I think with the way I've built this so far there might be some issues where if a field is missing during a PUT,
	//  like the name field; we might update the name to be blank?

	if err = repo.UpdateTransaction(c.getContext(ctx), bankAccountId, &transaction); err != nil {
		c.wrapPgError(ctx, err, "could not update transaction")
		return
	}

	balance, err := repo.GetBalances(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "could not get updated balances")
		return
	}

	c.getLog(ctx).Debugf("successfully updated transaction")

	ctx.JSON(map[string]interface{}{
		"transaction": transaction,
		"spending":    updatedExpenses,
		"balance":     balance,
	})
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

	isManual, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to validate if link is manual")
		return
	}

	if !isManual {
		c.returnError(ctx, http.StatusBadRequest, "cannot delete transactions for non-manual links")
		return
	}
}
