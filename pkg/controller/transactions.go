package controller

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"github.com/pkg/errors"
	"math"
	"net/http"
	"strings"
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

	if transaction.ExpenseId != nil && *transaction.ExpenseId > 0 {
		account, err := repo.GetAccount()
		if err != nil {
			c.wrapPgError(ctx, err, "could not get account to create transaction")
			return
		}

		expense, err := repo.GetExpense(bankAccountId, *transaction.ExpenseId)
		if err != nil {
			c.wrapPgError(ctx, err, "could not get expense provided for transaction")
			return
		}

		if err = c.addExpenseToTransaction(account, &transaction, expense); err != nil {
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to add expense to transaction")
			return
		}

		if err = repo.UpdateExpenses(bankAccountId, []models.Expense{
			*expense,
		}); err != nil {
			c.wrapPgError(ctx, err, "failed to update expense for transaction")
			return
		}
	}

	if err = repo.CreateTransaction(bankAccountId, &transaction); err != nil {
		c.wrapPgError(ctx, err, "could not create transaction")
		return
	}

	ctx.JSON(transaction)
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

	if err = c.processTransactionSpentFrom(repo, bankAccountId, &transaction, existingTransaction); err != nil {
		c.wrapPgError(ctx, err, "failed to process expense changes")
		return
	}

	// TODO Handle more complex transaction updates via the API.
	//  I think with the way I've built this so far there might be some issues where if a field is missing during a PUT,
	//  like the name field; we might update the name to be blank?

	if err = repo.UpdateTransaction(bankAccountId, &transaction); err != nil {
		c.wrapPgError(ctx, err, "could not update transaction")
		return
	}

	ctx.JSON(transaction)
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

func (c *Controller) processTransactionSpentFrom(
	repo repository.Repository,
	bankAccountId uint64,
	input, existing *models.Transaction,
) error {
	account, err := repo.GetAccount()
	if err != nil {
		return err
	}

	const (
		AddExpense = iota
		ChangeExpense
		RemoveExpense
	)

	var existingExpenseId uint64
	if existing.ExpenseId != nil {
		existingExpenseId = *existing.ExpenseId
	}

	var newExpenseId uint64
	if input.ExpenseId != nil {
		newExpenseId = *input.ExpenseId
	}

	var expensePlan int

	switch {
	case existingExpenseId == 0 && newExpenseId > 0:
		// Expense is being added to the transaction.
		expensePlan = AddExpense
	case existingExpenseId != 0 && newExpenseId != existingExpenseId && newExpenseId > 0:
		// Expense is being changed from one expense to another.
		expensePlan = ChangeExpense
	case existingExpenseId != 0 && newExpenseId == 0:
		// Expense is being removed from the transaction.
		expensePlan = RemoveExpense
	default:
		// TODO Handle transaction amount changes with expenses.
		return nil
	}

	// Retrieve the expenses that we need to work with and potentially update.
	var currentExpense, newExpense *models.Expense
	var currentErr, newErr error
	switch expensePlan {
	case AddExpense:
		newExpense, newErr = repo.GetExpense(bankAccountId, newExpenseId)
	case ChangeExpense:
		currentExpense, currentErr = repo.GetExpense(bankAccountId, existingExpenseId)
		newExpense, newErr = repo.GetExpense(bankAccountId, newExpenseId)
	case RemoveExpense:
		currentExpense, currentErr = repo.GetExpense(bankAccountId, existingExpenseId)
	}

	// If we failed to retrieve either of the expenses then something is wrong and we need to stop.
	switch {
	case currentErr != nil:
		return errors.Wrap(currentErr, "failed to retrieve the current expense for the transaction")
	case newErr != nil:
		return errors.Wrap(newErr, "failed to retrieve the new expense for the transaction")
	}

	expenseUpdates := make([]models.Expense, 0)

	switch expensePlan {
	case ChangeExpense, RemoveExpense:
		// If the transaction already has an expense then it should have an expense amount. If this is missing then
		// something is wrong.
		if existing.ExpenseAmount == nil {
			// TODO Handle missing expense amount when changing or removing a transaction's expense.
			panic("somethings wrong, expense amount missing")
		}

		// Add the amount we took from the expense back to it.
		currentExpense.CurrentAmount += *existing.ExpenseAmount

		// Now that we have added that money back to the expense we need to calculate the expense's next contribution.
		if err = currentExpense.CalculateNextContribution(
			account.Timezone,
			currentExpense.FundingSchedule.NextOccurrence,
			currentExpense.FundingSchedule.Rule,
		); err != nil {
			return errors.Wrap(err, "failed to calculate next contribution for current transaction expense")
		}

		// Then take all the fields that have changed and throw them in our list of things to update.
		expenseUpdates = append(expenseUpdates, *currentExpense)

		// If we are only removing the expense then we are done with this part.
		if expensePlan == RemoveExpense {
			break
		}

		// If we are changing the expense though then we want to fallthrough to handle the processing of the new
		// expense.
		fallthrough
	case AddExpense:
		if err = c.addExpenseToTransaction(account, input, newExpense); err != nil {
			return err
		}

		// Then take all the fields that have changed and throw them in our list of things to update.
		expenseUpdates = append(expenseUpdates, *newExpense)
	}

	return repo.UpdateExpenses(bankAccountId, expenseUpdates)
}

func (c *Controller) addExpenseToTransaction(
	account *models.Account,
	transaction *models.Transaction,
	expense *models.Expense,
) error {
	var allocationAmount int64
	// If the amount allocated to the expense we are adding to the transaction is less than the amount of the
	// transaction then we can only do a partial allocation.
	if expense.CurrentAmount < transaction.Amount {
		allocationAmount = expense.CurrentAmount
	} else {
		// Otherwise we will allocate the entire transaction amount from the expense.
		allocationAmount = transaction.Amount
	}

	// Subtract the amount we are taking from the expense from it's current amount.
	expense.CurrentAmount -= allocationAmount

	// Keep track of how much we took from the expense in case things change later.
	transaction.ExpenseAmount = &allocationAmount

	// Now that we have deducted the amount we need from the expense we need to recalculate it's next contribution.
	if err := expense.CalculateNextContribution(
		account.Timezone,
		expense.FundingSchedule.NextOccurrence,
		expense.FundingSchedule.Rule,
	); err != nil {
		return errors.Wrap(err, "failed to calculate next contribution for new transaction expense")
	}

	return nil
}
