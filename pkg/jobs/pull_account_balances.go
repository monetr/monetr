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

func (j *JobManager) EnqueuePullAccountBalances(job *work.Job) error {
	log := j.log.WithField("job", EnqueuePullAccountBalances)

	// We need an accountId, and all of the bank accounts for that account that can be updated.
	var accounts []struct {
		AccountID      uint64   `pg:"account_id"`
		BankAccountIDs []uint64 `pg:"bank_account_ids"`
	}

	// Query the database for all accounts with bank accounts that have a link type of plaid.
	_, err := j.db.Query(&accounts, `
		SELECT 
			"accounts"."account_id", 
			array_agg("bank_accounts"."bank_account_id") "bank_account_ids"
		FROM "accounts"
		INNER JOIN "bank_accounts" ON "bank_accounts"."account_id" = "accounts"."account_id"
		INNER JOIN "links" ON "links"."link_id" = "bank_account"."link_id" AND "links"."account_id" = "bank_account"."account_id"
		WHERE "links"."link_type" = ? -- We want to filter by link type so we don't try to update manual accounts.
		GROUP BY "accounts"."account_id"
	`, models.PlaidLinkType)
	if err != nil {
		err = errors.Wrap(err, "failed to retrieve accounts to update balances")
		log.WithError(err).Error("could not get accounts to update balances")
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

func (j *JobManager) PullAccountBalances(job *work.Job) error {
	return nil
}
