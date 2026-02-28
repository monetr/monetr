package repository

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10/orm"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type TransactionUpdateId struct {
	TransactionId ID[Transaction] `pg:"transaction_id"`
	BankAccountId ID[BankAccount] `pg:"bank_account_id"`
	Amount        int64           `pg:"amount"`
}

func (r *repositoryBase) InsertTransactions(ctx context.Context, transactions []Transaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	now := r.clock.Now()
	for i := range transactions {
		transactions[i].AccountId = r.AccountId()
		transactions[i].CreatedAt = now
	}
	_, err := r.txn.ModelContext(span.Context(), &transactions).Insert(&transactions)
	return errors.Wrap(err, "failed to insert transactions")
}

func (r *repositoryBase) GetTransactionsByPlaidId(
	ctx context.Context,
	linkId ID[Link],
	plaidTransactionIds []string,
) (map[string]Transaction, error) {
	if len(plaidTransactionIds) == 0 {
		return map[string]Transaction{}, nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"linkId":              linkId,
		"plaidTransactionIds": plaidTransactionIds,
	}

	items := make([]Transaction, 0)
	// Deliberatly include all transactions, regardless of delete status.
	// TODO This query is using a FROM for Transaction, but it would
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

	result := map[string]Transaction{}
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

func (r *repositoryBase) GetTransactionsByLunchFlowId(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	lunchFlowIds []string,
) (map[string]Transaction, error) {
	if len(lunchFlowIds) == 0 {
		return map[string]Transaction{}, nil
	}

	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	items := make([]Transaction, 0)
	// Deliberatly include all transactions, regardless of delete status.
	err := r.txn.ModelContext(span.Context(), &items).
		Relation("LunchFlowTransaction").
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		WhereIn(`"lunch_flow_transaction"."lunch_flow_id" IN (?)`, lunchFlowIds).
		Select(&items)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to retrieve transactions for lunch flow Ids")
	}

	span.Status = sentry.SpanStatusOK

	result := map[string]Transaction{}
	for i := range items {
		item := items[i]
		result[item.LunchFlowTransaction.LunchFlowId] = item
	}

	return result, nil
}

func (r *repositoryBase) GetTransactonsByUploadIdentifier(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	uploadIdentifiers []string,
) (map[string]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	items := make([]Transaction, 0)
	err := r.txn.ModelContext(span.Context(), &items).
		Where(`"account_id" = ?`, r.AccountId()).
		Where(`"bank_account_id" = ?`, bankAccountId).
		WhereIn(`"upload_identifier" IN (?)`, uploadIdentifiers).
		Select(&items)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retireve transactions by their upload identifier")
	}

	result := map[string]Transaction{}
	for i := range items {
		txn := items[i]
		result[*txn.UploadIdentifier] = txn
	}

	return result, nil
}

func (r *repositoryBase) GetTransactions(ctx context.Context, bankAccountId ID[BankAccount], limit, offset int) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"limit":         limit,
		"offset":        offset,
	}

	items := make([]Transaction, 0)
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

func (r *repositoryBase) GetPendingTransactions(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	limit, offset int,
) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"limit":         limit,
		"offset":        offset,
	}

	var items []Transaction
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

func (r *repositoryBase) GetTransactionsForSpending(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	spendingId ID[Spending],
	limit, offset int,
) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"accountId":     r.AccountId(),
		"bankAccountId": bankAccountId,
		"spendingId":    spendingId,
		"limit":         limit,
		"offset":        offset,
	}

	items := make([]Transaction, 0)
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

func (r *repositoryBase) GetTransaction(ctx context.Context, bankAccountId ID[BankAccount], transactionId ID[Transaction]) (*Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
		"bankAccountId": bankAccountId,
		"transactionId": transactionId,
	}

	var result Transaction
	err := r.txn.ModelContext(span.Context(), &result).
		Relation("LunchFlowTransaction").
		Relation("PlaidTransaction").
		Relation("PendingPlaidTransaction").
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

func (r *repositoryBase) CreateTransaction(ctx context.Context, bankAccountId ID[BankAccount], transaction *Transaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
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

func (r *repositoryBase) UpdateTransaction(ctx context.Context, bankAccountId ID[BankAccount], transaction *Transaction) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	span.Data = map[string]any{
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

// UpdateTransactions is unique in that it REQUIRES that all data on each
// transaction object be populated. It is doing a bulk update, so if data is
// missing it has the potential to overwrite a transaction incorrectly.
func (r *repositoryBase) UpdateTransactions(
	ctx context.Context,
	transactions []*Transaction,
) error {
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

func (r *repositoryBase) SoftDeleteTransaction(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionId ID[Transaction],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &Transaction{}).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."transaction_id" = ?`, transactionId).
		Set(`"deleted_at" = ?`, r.clock.Now().UTC()).
		Update()

	return errors.Wrap(err, "failed to soft-delete transaction")
}

func (r *repositoryBase) DeleteTransaction(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	transactionId ID[Transaction],
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	_, err := r.txn.ModelContext(span.Context(), &Transaction{}).
		Where(`"transaction"."account_id" = ?`, r.AccountId()).
		Where(`"transaction"."bank_account_id" = ?`, bankAccountId).
		Where(`"transaction"."transaction_id" = ?`, transactionId).
		ForceDelete()

	return errors.Wrap(err, "failed to delete transaction")
}

func (r *repositoryBase) GetTransactionsByPlaidTransactionId(
	ctx context.Context,
	linkId ID[Link],
	plaidTransactionIds []string,
) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]Transaction, 0)
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

func (r *repositoryBase) GetRecentDepositTransactions(
	ctx context.Context,
	bankAccountId ID[BankAccount],
) ([]Transaction, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	result := make([]Transaction, 0)
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

func (r *repositoryBase) ProcessTransactionSpentFrom(
	ctx context.Context,
	bankAccountId ID[BankAccount],
	input, existing *Transaction,
) (updatedExpenses []Spending, _ error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := r.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"bankAccountId": bankAccountId,
		"transactionId": existing.TransactionId,
	})

	account, err := r.GetAccount(span.Context())
	if err != nil {
		return nil, err
	}
	timezone, err := account.GetTimezone()
	if err != nil {
		return nil, err
	}

	const (
		AddExpense = iota
		ChangeExpense
		RemoveExpense
	)

	var existingSpendingId ID[Spending]
	if existing.SpendingId != nil {
		existingSpendingId = *existing.SpendingId
	}

	var newSpendingId ID[Spending]
	if input.SpendingId != nil {
		newSpendingId = *input.SpendingId
	}

	var expensePlan int

	switch {
	case existingSpendingId.IsZero() && !newSpendingId.IsZero():
		// Spending is being added to the transaction.
		expensePlan = AddExpense
		log = log.WithField("transactionAction", "AddExpense")
	case !existingSpendingId.IsZero() && newSpendingId != existingSpendingId && !newSpendingId.IsZero():
		// Spending is being changed from one expense to another.
		expensePlan = ChangeExpense
		log = log.WithField("transactionAction", "ChangeExpense")
	case !existingSpendingId.IsZero() && newSpendingId.IsZero():
		// Spending is being removed from the transaction.
		expensePlan = RemoveExpense
		log = log.WithField("transactionAction", "RemoveExpense")
	default:
		// TODO Handle transaction amount changes with expenses.
		return nil, nil
	}

	// Retrieve the expenses that we need to work with and potentially update.
	var currentExpense, newExpense *Spending
	var currentErr, newErr error
	switch expensePlan {
	case AddExpense:
		newExpense, newErr = r.GetSpendingById(
			span.Context(),
			bankAccountId,
			newSpendingId,
		)
	case ChangeExpense:
		currentExpense, currentErr = r.GetSpendingById(
			span.Context(),
			bankAccountId,
			existingSpendingId,
		)
		newExpense, newErr = r.GetSpendingById(
			span.Context(),
			bankAccountId,
			newSpendingId,
		)
	case RemoveExpense:
		currentExpense, currentErr = r.GetSpendingById(
			span.Context(),
			bankAccountId,
			existingSpendingId,
		)
	}

	// If we failed to retrieve either of the expenses then something is wrong and
	// we need to stop.
	switch {
	case currentErr != nil:
		return nil, errors.Wrap(currentErr, "failed to retrieve the current expense for the transaction")
	case newErr != nil:
		return nil, errors.Wrap(newErr, "failed to retrieve the new expense for the transaction")
	}

	expenseUpdates := make([]Spending, 0)

	switch expensePlan {
	case ChangeExpense, RemoveExpense:
		// If the transaction already has an expense then it should have an expense
		// amount. If this is missing then something is wrong.
		if existing.SpendingAmount == nil {
			// TODO Handle missing expense amount when changing or removing a
			// transaction's expense.
			panic("somethings wrong, expense amount missing")
		}

		// Add the amount we took from the expense back to it.
		currentExpense.CurrentAmount += *existing.SpendingAmount

		switch currentExpense.SpendingType {
		case SpendingTypeExpense:
		// Nothing special for expenses.
		case SpendingTypeGoal:
			// Revert the amount used for the current spending object.
			currentExpense.UsedAmount -= *existing.SpendingAmount
		}

		input.SpendingAmount = nil

		// Now that we have added that money back to the expense we need to
		// calculate the expense's next contribution.
		currentExpense.CalculateNextContribution(
			span.Context(),
			timezone,
			currentExpense.FundingSchedule,
			r.clock.Now(),
			log,
		)

		// Then take all the fields that have changed and throw them in our list of
		// things to update.
		expenseUpdates = append(expenseUpdates, *currentExpense)

		// If we are only removing the expense then we are done with this part.
		if expensePlan == RemoveExpense {
			break
		}

		// If we are changing the expense though then we want to fallthrough to
		// handle the processing of the new expense.
		fallthrough
	case AddExpense:
		if err = input.AddSpendingToTransaction(
			span.Context(),
			newExpense,
			timezone,
			r.clock.Now(),
			log,
		); err != nil {
			return nil, err
		}

		// Then take all the fields that have changed and throw them in our list of things to update.
		expenseUpdates = append(expenseUpdates, *newExpense)
	}

	return expenseUpdates, r.UpdateSpending(
		span.Context(),
		bankAccountId,
		expenseUpdates,
	)
}

func (r *repositoryBase) AddExpenseToTransaction(
	ctx context.Context,
	transaction *Transaction,
	spending *Spending,
) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	account, err := r.GetAccount(span.Context())
	if err != nil {
		return err
	}
	timezone, err := account.GetTimezone()
	if err != nil {
		return err
	}

	log := r.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"bankAccountId": transaction.BankAccountId,
		"transactionId": transaction.TransactionId,
		"spendingId":    spending.SpendingId,
	})

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
	case SpendingTypeExpense:
	// We don't need to do anything special if it's an expense, at least not right now.
	case SpendingTypeGoal:
		// Goals also keep track of how much has been spent, so increment the used amount.
		spending.UsedAmount += allocationAmount
	}

	// Keep track of how much we took from the spending in case things change later.
	transaction.SpendingAmount = &allocationAmount

	// Now that we have deducted the amount we need from the spending we need to recalculate it's next contribution.
	spending.CalculateNextContribution(
		span.Context(),
		timezone,
		spending.FundingSchedule,
		r.clock.Now(),
		log,
	)

	return nil
}
