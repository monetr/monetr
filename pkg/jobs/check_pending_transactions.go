package jobs

import (
	"github.com/gocraft/work"
)

const (
	EnqueueCheckPendingTransactions = "EnqueueCheckPendingTransactions"
	CheckPendingTransactions        = "CheckPendingTransactions"
)

type CheckingPendingTransactionsWorkItem struct {
	AccountID    uint64                   `pg:"account_id"`
	LinkID       uint64                   `pg:"link_id"`
	Transactions []PendingTransactionItem `pg:"transactions"`
}

type PendingTransactionItem struct {
	BankAccountID uint64 `pg:"bank_account_id"`
	TransactionID uint64 `pg:"transaction_id"`
}

func (j *jobManagerBase) getPlaidLinksWithPendingTransactions() ([]CheckingPendingTransactionsWorkItem, error) {
	return nil, nil
}

func (j *jobManagerBase) enqueueCheckPendingTransactions(job *work.Job) error {
	log := j.getLogForJob(job)

	accounts, err := j.getPlaidLinksWithPendingTransactions()
	if err != nil {
		log.WithError(err).Errorf("failed to retrieve links with pending transactions")
		return err
	}

	log.Infof("enqueueing %d account(s) for sync", len(accounts))

	return nil
}
