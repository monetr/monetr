package controller

import (
	"math"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	. "github.com/monetr/monetr/server/models"
	"github.com/sirupsen/logrus"
)

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
// @Failure 402 {object} SubscriptionNotActiveError The user's subscription is not active.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getTransactions(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	ctx.QueryParam("limit")

	limit := urlParamIntDefault(ctx, "limit", 25)
	offset := urlParamIntDefault(ctx, "offset", 0)

	if limit < 1 {
		return c.badRequest(ctx, "limit must be at least 1")
	} else if limit > 100 {
		return c.badRequest(ctx, "limit cannot be greater than 100")
	}

	if offset < 0 {
		return c.badRequest(ctx, "offset cannot be less than 0")
	}

	// Only let a maximum of 100 transactions be requested at a time.
	limit = int(math.Min(100, float64(limit)))

	repo := c.mustGetAuthenticatedRepository(ctx)

	transactions, err := repo.GetTransactions(c.getContext(ctx), bankAccountId, limit, offset)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transactions")
	}

	return ctx.JSON(http.StatusOK, transactions)
}

// getTransactionById will simply return a single transaction for the given bank
// and transaction specified. If the transaction does not exist then a 404 not
// found will be returned via the wrapPgError.
func (c *Controller) getTransactionById(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionId, err := ParseID[Transaction](ctx.Param("transactionId"))
	if err != nil || transactionId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	transaction, err := repo.GetTransaction(
		c.getContext(ctx),
		bankAccountId,
		transactionId,
	)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve transaction")
	}

	return ctx.JSON(http.StatusOK, transaction)
}

func (c *Controller) getSimilarTransactionsById(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionId, err := ParseID[Transaction](ctx.Param("transactionId"))
	if err != nil || transactionId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	cluster, _ := repo.GetTransactionClusterByMember(
		c.getContext(ctx),
		bankAccountId,
		transactionId,
	)

	// If there are no similar transactions then return no content, this will
	// prevent react-query from retrying in a weird way.
	if cluster == nil || len(cluster.Members) == 0 {
		return ctx.NoContent(http.StatusNoContent)
	}

	return ctx.JSON(http.StatusOK, cluster)
}

func (c *Controller) postTransactions(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	isManual, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to validate if link is manual")
	}

	if !isManual {
		return c.badRequest(ctx, "cannot create transactions for non-manual links")
	}

	var transaction Transaction
	if err = ctx.Bind(&transaction); err != nil {
		return c.invalidJson(ctx)
	}

	transaction.TransactionId = ""
	transaction.BankAccountId = bankAccountId
	transaction.Name = strings.TrimSpace(transaction.Name)
	transaction.MerchantName = strings.TrimSpace(transaction.MerchantName)
	transaction.OriginalName = transaction.Name

	if transaction.Name == "" {
		return c.badRequest(ctx, "transaction must have a name")
	}

	if transaction.Amount <= 0 {
		return c.badRequest(ctx, "transaction amount must be greater than 0")
	}

	var updatedSpending *Spending
	if transaction.SpendingId != nil && !(*transaction.SpendingId).IsZero() {
		updatedSpending, err = repo.GetSpendingById(
			c.getContext(ctx),
			bankAccountId,
			*transaction.SpendingId,
		)
		if err != nil {
			return c.wrapPgError(ctx, err, "could not get spending provided for transaction")
		}

		if err = repo.AddExpenseToTransaction(
			c.getContext(ctx),
			&transaction,
			updatedSpending,
		); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to add expense to transaction")
		}

		if err = repo.UpdateSpending(c.getContext(ctx), bankAccountId, []Spending{
			*updatedSpending,
		}); err != nil {
			return c.wrapPgError(ctx, err, "failed to update spending for transaction")
		}
	}

	if err = repo.CreateTransaction(c.getContext(ctx), bankAccountId, &transaction); err != nil {
		return c.wrapPgError(ctx, err, "could not create transaction")
	}

	returnedObject := map[string]interface{}{
		"transaction": transaction,
	}

	// If an expense was updated as part of this transaction being created then we want to include that updated expense
	// in our response so the UI can update its redux store.
	if updatedSpending != nil {
		returnedObject["spending"] = *updatedSpending
	}

	return ctx.JSON(http.StatusOK, returnedObject)
}

func (c *Controller) putTransactions(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionId, err := ParseID[Transaction](ctx.Param("transactionId"))
	if err != nil || transactionId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction Id")
	}

	var transaction Transaction
	if err := ctx.Bind(&transaction); err != nil {
		return c.invalidJson(ctx)
	}

	transaction.TransactionId = transactionId
	transaction.BankAccountId = bankAccountId

	repo := c.mustGetAuthenticatedRepository(ctx)

	isManual, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to validate if link is manual")
	}

	existingTransaction, err := repo.GetTransaction(c.getContext(ctx), bankAccountId, transactionId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve existing transaction for update")
	}

	if transaction.IsAddition() && transaction.SpendingId != nil {
		return c.badRequest(ctx, "cannot specify a spent from on a deposit")
	}

	transaction.PlaidTransactionId = existingTransaction.PlaidTransactionId
	transaction.PendingPlaidTransactionId = existingTransaction.PendingPlaidTransactionId

	if !isManual {
		// Prevent the user from attempting to change a transaction's amount if we are on a plaid link.
		if existingTransaction.Amount != transaction.Amount {
			return c.badRequest(ctx, "cannot change transaction amount on non-manual links")
		}

		if existingTransaction.IsPending != transaction.IsPending {
			return c.badRequest(ctx, "cannot change transaction pending state on non-manual links")
		}

		if !existingTransaction.Date.Equal(transaction.Date) {
			c.getLog(ctx).WithFields(logrus.Fields{
				"existingDate": existingTransaction.Date,
				"newDate":      transaction.Date,
			}).Warn("cannot change transaction date on non-manual links")
			return c.badRequest(ctx, "cannot change transaction date on non-manual links")
		}

		transaction.OriginalName = existingTransaction.OriginalName
		transaction.OriginalMerchantName = existingTransaction.OriginalMerchantName
	}

	updatedExpenses, err := repo.ProcessTransactionSpentFrom(c.getContext(ctx), bankAccountId, &transaction, existingTransaction)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to process expense changes")
	}

	// TODO Handle more complex transaction updates via the API.
	//  I think with the way I've built this so far there might be some issues where if a field is missing during a PUT,
	//  like the name field; we might update the name to be blank?

	if err = repo.UpdateTransaction(c.getContext(ctx), bankAccountId, &transaction); err != nil {
		return c.wrapPgError(ctx, err, "could not update transaction")
	}

	balance, err := repo.GetBalances(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "could not get updated balances")
	}

	c.getLog(ctx).Debugf("successfully updated transaction")

	result := map[string]interface{}{
		"transaction": transaction,
		"balance":     balance,
	}

	if updatedExpenses != nil {
		result["spending"] = updatedExpenses
	}

	return ctx.JSON(http.StatusOK, result)
}

func (c *Controller) deleteTransactions(ctx echo.Context) error {
	bankAccountId, err := ParseID[BankAccount](ctx.Param("bankAccountId"))
	if err != nil || bankAccountId.IsZero() {
		return c.badRequest(ctx, "must specify a valid bank account Id")
	}

	transactionId, err := ParseID[Transaction](ctx.Param("transactionId"))
	if err != nil || transactionId.IsZero() {
		return c.badRequest(ctx, "must specify a valid transaction Id")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	isManual, err := repo.GetLinkIsManualByBankAccountId(c.getContext(ctx), bankAccountId)
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to validate if link is manual")
	}

	if !isManual {
		return c.returnError(ctx, http.StatusBadRequest, "cannot delete transactions for non-manual links")
	}

	return ctx.NoContent(http.StatusOK)
}
