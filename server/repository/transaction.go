package repository

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10/orm"
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

	items := make([]models.Transaction, 0)
	// Deliberatly include all transactions, regardless of delete status.
	// TODO This query is using a FROM for models.Transaction, but it would
	// probably be more efficient to use the plaid transactions table as the base
	// and then join on transaction. But for now this is still fine.
	err := r.txn.ModelContext(span.Context(), &items).
		Relation("PlaidTransaction").
		Relation("PendingPlaidTransaction").
		Join(`INNER JOIN "bank_accounts" AS "bank_account"`).
		JoinOn(`"bank_account"."bank_account_id" = "transaction"."bank_account_id" AND "bank_account"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ?`, linkId).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereIn(`"plaid_transaction"."plaid_id" IN (?)`, plaidTransactionIds).
				WhereInOr(`"pending_plaid_transaction"."plaid_id" IN (?)`, plaidTransactionIds)
			return q, nil
		}).
		Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transaction Ids for plaid Ids")
	}

	span.Status = sentry.SpanStatusOK

	result := map[string]models.Transaction{}
	for i := range items {
		item := items[i]
		if item.PlaidTransaction != nil {
			result[item.PlaidTransaction.PlaidId] = item
		}
		if item.PendingPlaidTransaction != nil {
			result[item.PendingPlaidTransaction.PlaidId] = item
		}
	}

	return result, nil
}

// GetTransactionsByTellerId is used by the sync code to compare transactions
// from teller to the transactions in the database. All transactions that have
// the specified teller Ids will be included if they exist. But if
// includePending is true then additional transactions might also be returned.
// This way pending transactios that are no longer present in the API can be
// compared easily.
func (r *repositoryBase) GetTransactionsByTellerId(
	ctx context.Context,
	bankAccountId uint64,
	tellerIds []string,
	includePending bool,
) ([]models.Transaction, error) {
	if len(tellerIds) == 0 {
		return []models.Transaction{}, nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"bankAccountId": bankAccountId,
		"tellerIds":     tellerIds,
	}

	items := make([]models.Transaction, 0)
	query := r.txn.ModelContext(span.Context(), &items).
		Relation("TellerTransaction").
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId)

	if includePending {
		query = query.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereInOr(`"teller_transaction"."teller_id" IN (?)`, tellerIds).
				WhereOr(`"teller_transaction"."is_pending" = ?`, true)
			return q, nil
		})
	} else {
		query = query.WhereIn(`"teller_transaction"."teller_id" IN (?)`, tellerIds)
	}

	err := query.Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to find transactions by teller Id")
	}

	return items, nil
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

	items := make([]models.Transaction, 0)
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

func (r *repositoryBase) GetTransactionsAfter(ctx context.Context, bankAccountId uint64, after *time.Time) ([]models.Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]interface{}{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"after":         after,
	}

	var items []models.Transaction
	query := r.txn.ModelContext(span.Context(), &items).
		Relation("TellerTransaction").
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Order(`date DESC`).
		Order(`transaction_id DESC`)

	if after != nil {
		query = query.Where(`"transaction"."date" >= ?`, *after)
	}

	err := query.Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, crumbs.WrapError(span.Context(), err, "failed to retrieve transactions")
	}

	span.Status = sentry.SpanStatusOK

	return items, nil
}

func (r *repositoryBase) GetPendingTransactions(
	ctx context.Context,
	bankAccountId uint64,
	limit, offset int,
) ([]models.Transaction, error) {
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
		Where(`"transaction"."is_pending" = ?`, true).
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
		Relation("PlaidTransaction").
		Relation("PendingPlaidTransaction").
		Relation("TellerTransaction").
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
		Set(`"deleted_at" = ?`, r.clock.Now().UTC()).
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
		Join(`INNER JOIN "plaid_transactions" AS "plaid_transaction"`).
		JoinOn(`"plaid_transaction"."plaid_transaction_id" IN ("transaction"."plaid_transaction_id", "transaction"."pending_plaid_transaction_id") AND "plaid_transaction"."account_id" = "transaction"."account_id"`).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"bank_account"."link_id" = ?`, linkId).
		WhereIn(`"plaid_transaction"."plaid_id" IN (?)`, plaidTransactionIds).
		DistinctOn(`"transaction"."transaction_id"`).
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
