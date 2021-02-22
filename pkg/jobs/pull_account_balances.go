package jobs

import (
	"github.com/gocraft/work"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/pkg/errors"
)

const (
	EnqueuePullAccountBalances = "EnqueuePullAccountBalances"
	PullAccountBalances        = "PullAccountBalances"
)

type PullAccountBalanceWorkItem struct {
	AccountID      uint64   `pg:"account_id"`
	BankAccountIDs []uint64 `pg:"bank_account_ids,type:bigint[]"`
}

func (j *jobManagerBase) getPlaidBankAccountsByAccount() ([]PullAccountBalanceWorkItem, error) {
	// We need an accountId, and all of the bank accounts for that account that can be updated.
	var accounts []PullAccountBalanceWorkItem

	// Query the database for all accounts with bank accounts that have a link type of plaid.
	_, err := j.db.Query(&accounts, `
		SELECT 
			"accounts"."account_id", 
			array_agg("bank_accounts"."bank_account_id") "bank_account_ids"
		FROM "accounts"
		INNER JOIN "bank_accounts" ON "bank_accounts"."account_id" = "accounts"."account_id"
		INNER JOIN "links" ON "links"."link_id" = "bank_accounts"."link_id" AND "links"."account_id" = "bank_accounts"."account_id"
		WHERE "links"."link_type" = ?
		GROUP BY "accounts"."account_id"
	`, models.PlaidLinkType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve accounts to update balances")
	}

	return accounts, nil
}

func (j *jobManagerBase) enqueuePullAccountBalances(job *work.Job) error {
	log := j.getLogForJob(job)

	accounts, err := j.getPlaidBankAccountsByAccount()
	if err != nil {
		log.WithError(err).Errorf("failed to retrieve bank accounts that need to by synced")
		return err
	}

	log.Infof("enqueueing %d account(s) for sync", len(accounts))

	for _, account := range accounts {
		accountLog := log.WithField("accountId", account.AccountID)
		accountLog.Trace("enqueueing for account balance update")
		_, err := j.queue.EnqueueUnique(PullAccountBalances, map[string]interface{}{
			"accountId":      account.AccountID,
			"bankAccountIds": account.BankAccountIDs,
		})
		if err != nil {
			err = errors.Wrap(err, "failed to enqueue account")
			accountLog.WithError(err).Error("could not enqueue account, data will not be synced")
			continue
		}
		accountLog.Trace("successfully enqueued account for account balance update")
	}

	return nil
}

func (j *jobManagerBase) pullAccountBalances(job *work.Job) error {
	log := j.getLogForJob(job)
	log.Infof("pulling account balances")
	return nil
}
