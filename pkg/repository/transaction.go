package repository

import (
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
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

func (r *repositoryBase) GetTransactionsByPlaidId(linkId uint64, plaidTransactionIds []string) (map[string]TransactionUpdateId, error) {
	type Transaction struct {
		tableName          string `pg:"transactions"`
		PlaidTransactionId string `pg:"plaid_transaction_id"`
		TransactionUpdateId
	}
	var items []Transaction
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

	result := map[string]TransactionUpdateId{}
	for _, item := range items {
		result[item.PlaidTransactionId] = item.TransactionUpdateId
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
