package jobs

import (
	"github.com/gocraft/work"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	EnqueueCheckPendingTransactions = "EnqueueCheckPendingTransactions"
	CheckPendingTransactions        = "CheckPendingTransactions"
)

type CheckingPendingTransactionsWorkItem struct {
	tableName    string                   `pg:"links"`
	AccountID    uint64                   `pg:"account_id"`
	LinkID       uint64                   `pg:"link_id"`
	Transactions []PendingTransactionItem `pg:"transactions"`
}

type PendingTransactionItem struct {
	BankAccountID uint64 `pg:"bank_account_id"`
	TransactionID uint64 `pg:"transaction_id"`
}

func (j *jobManagerBase) enqueueCheckPendingTransactions(job *work.Job) error {
	log := j.getLogForJob(job)

	var workItems []repository.CheckingPendingTransactionsItem
	var err error
	if err = j.getJobHelperRepository(job, func(repo repository.JobRepository) error {
		workItems, err = repo.GetBankAccountsWithPendingTransactions()
		return err
	}); err != nil {
		log.WithError(err).Errorf("failed to get bank accounts with pending transactions")
		return err
	}

	log.Infof("enqueueing %d link(s) for sync", len(workItems))

	for _, item := range workItems {
		itemLog := log.WithFields(logrus.Fields{
			"accountId": item.AccountId,
			"linkId":    item.LinkId,
		})
		itemLog.Trace("enqueueing for pending transaction processing")
		_, err = j.queue.EnqueueUnique(CheckPendingTransactions, map[string]interface{}{
			"accountId": item.AccountId,
			"linkId":    item.LinkId,
		})
		if err != nil {
			err = errors.Wrap(err, "failed to enqueue pending transactions work item")
			itemLog.WithError(err).Error("could not enqueue link, pending transactions will not be checked")
			continue
		}
		itemLog.Trace("successfully enqueued account for pending transaction check")
	}

	return nil
}

func (j *jobManagerBase) checkPendingTransactions(job *work.Job) error {
	return nil
}