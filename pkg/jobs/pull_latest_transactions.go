package jobs

import (
	"github.com/gocraft/work"
	"github.com/pkg/errors"
)

const (
	EnqueuePullLatestTransactions = "EnqueuePullLatestTransactions"
	PullLatestTransactions        = "PullLatestTransactions"
)

func (j *jobManagerBase) enqueuePullLatestTransactions(job *work.Job) error {
	log := j.getLogForJob(job)

	accounts, err := j.getPlaidLinksByAccount()
	if err != nil {
		log.WithError(err).Errorf("failed to retrieve bank accounts that need to by synced")
		return err
	}

	log.Infof("enqueing %d account(s) to pull latest transactions", len(accounts))

	for _, account := range accounts {
		accountLog := log.WithField("accountId", account.AccountID)
		accountLog.Trace("enqueueing for latest transactions update")
		_, err := j.queue.EnqueueUnique(PullLatestTransactions, map[string]interface{}{
			"accountId": account.AccountID,
			// TODO (elliotcourant) Convert pull latest transactions to use linkIds instead.
		})
		if err != nil {
			err = errors.Wrap(err, "failed to enqueue account")
			accountLog.WithError(err).Error("could not enqueue account, data will not be synced")
			continue
		}
		accountLog.Trace("successfully enqueued account for latest transactions update")
	}

	return nil
}

func (j *jobManagerBase) pullLatestTransactions(job *work.Job) error {
	return nil
}
