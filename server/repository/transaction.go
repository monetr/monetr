package repository

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

type TransactionUpdateId struct {
	TransactionId uint64 `pg:"transaction_id"`
	BankAccountId uint64 `pg:"bank_account_id"`
	Amount        int64  `pg:"amount"`
}

func (r *repositoryBase) InsertTransactions(ctx context.Context, transactions []models.Transaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
	}
	_, err := r.txn.ModelContext(span.Context(), &transactions).Insert(&transactions)
	return errors.Wrap(err, "failed to insert transactions")
}

func (r *repositoryBase) GetTransactionsByPlaidId(ctx context.Context, linkId uint64, plaidTransactionIds []string) (map[string]models.Transaction, error) {
	if len(plaidTransactionIds) == 0 {
		return map[string]models.Transaction{}, nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"linkId":              linkId,
		"plaidTransactionIds": plaidTransactionIds,
	}

	var items []models.Transaction
	// Deliberatly include all transactions, regardless of delete status.
	err := r.txn.ModelContext(span.Context(), &items).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."bank_account_id" = "transaction"."bank_account_id" AND "bank_account"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ?`, linkId).
		WhereIn(`"transaction"."plaid_transaction_id" IN (?)`, plaidTransactionIds).
		Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction Ids for plaid Ids")
	}

	span.Status = sentry.SpanStatusOK

	result := map[string]models.Transaction{}
	for _, item := range items {
		result[item.PlaidTransactionId] = item
	}

	return result, nil
}

func (r *repositoryBase) GetTransactions(ctx context.Context, bankAccountId uint64, limit, offset int) ([]models.Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"limit":         limit,
		"offset":        offset,
	}

	var items []models.Transaction
	err := r.txn.ModelContext(span.Context(), &items).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."deleted_at" IS NULL`).
		Limit(limit).
		Offset(offset).
		Order(`date DESC`).
		Order(`transaction_id DESC`).
		Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(span.Context(), err, "failed to retrieve transactions")
	}

	span.Status = sentry.SpanStatusOK

	return items, nil
}

func (r *repositoryBase) GetTransactionsForSpending(ctx context.Context, bankAccountId, spendingId uint64, limit, offset int) ([]models.Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingId":    spendingId,
		"limit":         limit,
		"offset":        offset,
	}

	var items []models.Transaction
	err := r.txn.ModelContext(span.Context(), &items).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."spending_id" = ?`, spendingId).
		Where(`"transaction"."deleted_at" IS NULL`).
		Limit(limit).
		Offset(offset).
		Order(`date DESC`).
		Order(`transaction_id DESC`).
		Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transactions for spending")
	}

	span.Status = sentry.SpanStatusOK

	return items, nil
}

func (r *repositoryBase) GetTransaction(ctx context.Context, bankAccountId, transactionId uint64) (*models.Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"bankAccountId": bankAccountId,
		"transactionId": transactionId,
	}

	var result models.Transaction
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."transaction_id" = ?`, transactionId).
		Select(&result)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction")
	}

	span.Status = sentry.SpanStatusOK

	return &result, nil
}

func (r *repositoryBase) CreateTransaction(ctx context.Context, bankAccountId uint64, transaction *models.Transaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"bankAccountId": bankAccountId,
	}

	transaction.AccountId = r.AccountId()
	transaction.BankAccountId = bankAccountId

	_, err := r.txn.ModelContext(span.Context(), transaction).Insert(transaction)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to create transaction")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

func (r *repositoryBase) UpdateTransaction(ctx context.Context, bankAccountId uint64, transaction *models.Transaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"bankAccountId": bankAccountId,
	}

	transaction.AccountId = r.AccountId()

	_, err := r.txn.ModelContext(span.Context(), transaction).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		WherePK().
		Update(&transaction)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return errors.Wrap(err, "failed to update transaction")
	}

	span.Status = sentry.SpanStatusOK

	return nil
}

// UpdateTransactions is unique in that it REQUIRES that all data on each transaction object be populated. It is doing a
// bulk update, so if data is missing it has the potential to overwrite a transaction incorrectly.
func (r *repositoryBase) UpdateTransactions(ctx context.Context, transactions []*models.Transaction) error {
	span := crumbs.StartFnTrace(ctx)
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

func (r *repositoryBase) DeleteTransaction(ctx context.Context, bankAccountId, transactionId uint64) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &models.Transaction{}).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."transaction_id" = ?`, transactionId).
		Set(`"deleted_at" = ?`, time.Now().UTC()).
		Update()

	return errors.Wrap(err, "failed to soft-delete transaction")
}

func (r *repositoryBase) GetTransactionsByPlaidTransactionId(ctx context.Context, linkId uint64, plaidTransactionIds []string) ([]models.Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]models.Transaction, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."bank_account_id" = "transaction"."bank_account_id" AND "bank_account"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ?`, linkId).
		WhereIn(`"transaction"."plaid_transaction_id" IN (?)`, plaidTransactionIds).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve transactions by plaid Id")
	}

	return result, nil
}

func (r *repositoryBase) GetRecentDepositTransactions(ctx context.Context, bankAccountId uint64) ([]models.Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]models.Transaction, 0)
	err := r.txn.ModelContext(span.Context(), &result).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."amount" < 0`). // Negative transactions are deposits.
		Where(`"transaction"."date" >= ?`, time.Now().Add(-24*time.Hour)).
		Where(`"transaction"."deleted_at" IS NULL`).
		Select(&result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve recent deposit transactions")
	}

	return result, nil
}

func (r *repositoryBase) ProcessTransactionSpentFrom(ctx context.Context, bankAccountId uint64, input, existing *models.Transaction) (updatedExpenses []models.Spending, _ error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	account, err := r.GetAccount(span.Context())
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
		newExpense, newErr = r.GetSpendingById(span.Context(), bankAccountId, newSpendingId)
	case ChangeExpense:
		currentExpense, currentErr = r.GetSpendingById(span.Context(), bankAccountId, existingSpendingId)
		newExpense, newErr = r.GetSpendingById(span.Context(), bankAccountId, newSpendingId)
	case RemoveExpense:
		currentExpense, currentErr = r.GetSpendingById(span.Context(), bankAccountId, existingSpendingId)
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
			span.Context(),
			account.Timezone,
			currentExpense.FundingSchedule,
			time.Now(),
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
		if err = input.AddSpendingToTransaction(span.Context(), newExpense, account); err != nil {
			return nil, err
		}

		// Then take all the fields that have changed and throw them in our list of things to update.
		expenseUpdates = append(expenseUpdates, *newExpense)
	}

	return expenseUpdates, r.UpdateSpending(span.Context(), bankAccountId, expenseUpdates)
}

func (r *repositoryBase) AddExpenseToTransaction(ctx context.Context, transaction *models.Transaction, spending *models.Spending) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	account, err := r.GetAccount(span.Context())
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
	if err = spending.CalculateNextContribution(
		span.Context(),
		account.Timezone,
		spending.FundingSchedule,
		time.Now(),
	); err != nil {
		return errors.Wrap(err, "failed to calculate next contribution for new transaction expense")
	}

	return nil
}
