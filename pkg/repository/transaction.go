package repository

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/pkg/errors"
)

type TransactionUpdateId struct {
	TransactionId uint64 `pg:"transaction_id"`
	BankAccountId uint64 `pg:"bank_account_id"`
	Amount        int64  `pg:"amount"`
}

func (r *repositoryBase) InsertTransactions(transactions []models.Transaction) error {
	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
	}
	_, err := r.txn.Model(&transactions).Insert(&transactions)
	return errors.Wrap(err, "failed to insert transactions")
}

func (r *repositoryBase) GetPendingTransactionsForBankAccount(bankAccountId uint64) ([]models.Transaction, error) {
	var result []models.Transaction
	err := r.txn.Model(&result).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."is_pending" = ?`, true).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve pending transactions for bank account")
	}

	return result, nil
}

func (r *repositoryBase) GetTransactionsByPlaidId(linkId uint64, plaidTransactionIds []string) (map[string]models.Transaction, error) {
	var items []models.Transaction
	err := r.txn.Model(&items).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."bank_account_id" = "transaction"."bank_account_id" AND "bank_account"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ?`, linkId).
		WhereIn(`"transaction"."plaid_transaction_id" IN (?)`, plaidTransactionIds).
		Select(&items)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve transaction Ids for plaid Ids")
	}

	result := map[string]models.Transaction{}
	for _, item := range items {
		result[item.PlaidTransactionId] = item
	}

	return result, nil
}

func (r *repositoryBase) GetTransactions(bankAccountId uint64, limit, offset int) ([]models.Transaction, error) {
	var items []models.Transaction
	err := r.txn.Model(&items).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Limit(limit).
		Offset(offset).
		Order(`date DESC`).
		Order(`transaction_id DESC`).
		Select(&items)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve transactions")
	}

	return items, nil
}

func (r *repositoryBase) GetTransaction(bankAccountId, transactionId uint64) (*models.Transaction, error) {
	var result models.Transaction
	err := r.txn.Model(&result).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."transaction_id" = ?`, transactionId).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve transaction")
	}

	return &result, nil
}

func (r *repositoryBase) CreateTransaction(bankAccountId uint64, transaction *models.Transaction) error {
	transaction.AccountId = r.AccountId()
	transaction.BankAccountId = bankAccountId

	_, err := r.txn.Model(transaction).Insert(transaction)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}

	return nil
}

func (r *repositoryBase) UpdateTransaction(bankAccountId uint64, transaction *models.Transaction) error {
	transaction.AccountId = r.AccountId()

	_, err := r.txn.Model(transaction).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		WherePK().
		Update(&transaction)
	if err != nil {
		return errors.Wrap(err, "failed to update transaction")
	}

	return nil
}

// UpdateTransactions is unique in that it REQUIRES that all data on each transaction object be populated. It is doing a
// bulk update, so if data is missing it has the potential to overwrite a transaction incorrectly.
func (r *repositoryBase) UpdateTransactions(ctx context.Context, transactions []*models.Transaction) error {
	span := sentry.StartSpan(ctx, "Update Transactions")
	defer span.Finish()

	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
	}

	result, err := r.txn.ModelContext(span.Context(), &transactions).
		WherePK().
		Update(&transactions)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update transactions")
	}

	if affected := result.RowsAffected(); affected != len(transactions) {
		span.Status = sentry.SpanStatusDataLoss
		return errors.Errorf("not all transactions updated, expected: %d updated: %d", len(transactions), affected)
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) DeleteTransaction(bankAccountId, transactionId uint64) error {
	_, err := r.txn.Model(&models.Transaction{}).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."transaction_id" = ?`, transactionId).
		Delete()

	return errors.Wrap(err, "failed to delete transaction")
}

func (r *repositoryBase) GetTransactionsByPlaidTransactionId(linkId uint64, plaidTransactionIds []string) ([]models.Transaction, error) {
	result := make([]models.Transaction, 0)
	err := r.txn.Model(&result).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."bank_account_id" = "transaction"."bank_account_id" AND "bank_account"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ?`, linkId).
		Where(`"transaction"."plaid_transaction_id" IN (?)`, plaidTransactionIds).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve transactions by plaid Id")
	}

	return result, nil
}

func (r *repositoryBase) ProcessTransactionSpentFrom(bankAccountId uint64, input, existing *models.Transaction) (updatedExpenses []models.Spending, _ error) {
	account, err := r.GetAccount()
	if err != nil {
		return nil, err
	}

	const (
		AddExpense = iota
		ChangeExpense
		RemoveExpense
	)

	var existingSpendingId uint64
	if existing.SpendingId != nil {
		existingSpendingId = *existing.SpendingId
	}

	var newSpendingId uint64
	if input.SpendingId != nil {
		newSpendingId = *input.SpendingId
	}

	var expensePlan int

	switch {
	case existingSpendingId == 0 && newSpendingId > 0:
		// Spending is being added to the transaction.
		expensePlan = AddExpense
	case existingSpendingId != 0 && newSpendingId != existingSpendingId && newSpendingId > 0:
		// Spending is being changed from one expense to another.
		expensePlan = ChangeExpense
	case existingSpendingId != 0 && newSpendingId == 0:
		// Spending is being removed from the transaction.
		expensePlan = RemoveExpense
	default:
		// TODO Handle transaction amount changes with expenses.
		return nil, nil
	}

	// Retrieve the expenses that we need to work with and potentially update.
	var currentExpense, newExpense *models.Spending
	var currentErr, newErr error
	switch expensePlan {
	case AddExpense:
		newExpense, newErr = r.GetSpendingById(bankAccountId, newSpendingId)
	case ChangeExpense:
		currentExpense, currentErr = r.GetSpendingById(bankAccountId, existingSpendingId)
		newExpense, newErr = r.GetSpendingById(bankAccountId, newSpendingId)
	case RemoveExpense:
		currentExpense, currentErr = r.GetSpendingById(bankAccountId, existingSpendingId)
	}

	// If we failed to retrieve either of the expenses then something is wrong and we need to stop.
	switch {
	case currentErr != nil:
		return nil, errors.Wrap(currentErr, "failed to retrieve the current expense for the transaction")
	case newErr != nil:
		return nil, errors.Wrap(newErr, "failed to retrieve the new expense for the transaction")
	}

	expenseUpdates := make([]models.Spending, 0)

	switch expensePlan {
	case ChangeExpense, RemoveExpense:
		// If the transaction already has an expense then it should have an expense amount. If this is missing then
		// something is wrong.
		if existing.SpendingAmount == nil {
			// TODO Handle missing expense amount when changing or removing a transaction's expense.
			panic("somethings wrong, expense amount missing")
		}

		// Add the amount we took from the expense back to it.
		currentExpense.CurrentAmount += *existing.SpendingAmount

		switch currentExpense.SpendingType {
		case models.SpendingTypeExpense:
		// Nothing special for expenses.
		case models.SpendingTypeGoal:
			// Revert the amount used for the current spending object.
			currentExpense.UsedAmount -= *existing.SpendingAmount
		}

		input.SpendingAmount = nil

		// Now that we have added that money back to the expense we need to calculate the expense's next contribution.
		if err = currentExpense.CalculateNextContribution(
			account.Timezone,
			currentExpense.FundingSchedule.NextOccurrence,
			currentExpense.FundingSchedule.Rule,
		); err != nil {
			return nil, errors.Wrap(err, "failed to calculate next contribution for current transaction expense")
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
		if err = r.AddExpenseToTransaction(input, newExpense); err != nil {
			return nil, err
		}

		// Then take all the fields that have changed and throw them in our list of things to update.
		expenseUpdates = append(expenseUpdates, *newExpense)
	}

	return expenseUpdates, r.UpdateExpenses(bankAccountId, expenseUpdates)
}

func (r *repositoryBase) AddExpenseToTransaction(transaction *models.Transaction, spending *models.Spending) error {
	account, err := r.GetAccount()
	if err != nil {
		return err
	}

	var allocationAmount int64
	// If the amount allocated to the spending we are adding to the transaction is less than the amount of the
	// transaction then we can only do a partial allocation.
	if spending.CurrentAmount < transaction.Amount {
		allocationAmount = spending.CurrentAmount
	} else {
		// Otherwise we will allocate the entire transaction amount from the spending.
		allocationAmount = transaction.Amount
	}

	// Subtract the amount we are taking from the spending from it's current amount.
	spending.CurrentAmount -= allocationAmount

	switch spending.SpendingType {
	case models.SpendingTypeExpense:
	// We don't need to do anything special if it's an expense, at least not right now.
	case models.SpendingTypeGoal:
		// Goals also keep track of how much has been spent, so increment the used amount.
		spending.UsedAmount += allocationAmount
	}

	// Keep track of how much we took from the spending in case things change later.
	transaction.SpendingAmount = &allocationAmount

	// Now that we have deducted the amount we need from the spending we need to recalculate it's next contribution.
	if err := spending.CalculateNextContribution(
		account.Timezone,
		spending.FundingSchedule.NextOccurrence,
		spending.FundingSchedule.Rule,
	); err != nil {
		return errors.Wrap(err, "failed to calculate next contribution for new transaction expense")
	}

	return nil
}