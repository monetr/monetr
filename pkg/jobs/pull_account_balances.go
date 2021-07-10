package jobs

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"

	"github.com/gocraft/work"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/plaid/plaid-go/plaid"
)

const (
	EnqueuePullAccountBalances = "EnqueuePullAccountBalances"
	PullAccountBalances        = "PullAccountBalances"
)

type PullAccountBalanceWorkItem struct {
	AccountID uint64   `pg:"account_id"`
	LinkIDs   []uint64 `pg:"link_ids,type:bigint[]"`
}

func (j *jobManagerBase) getPlaidLinksByAccount() ([]PullAccountBalanceWorkItem, error) {
	// We need an accountId, and all of the bank accounts for that account that can be updated.
	var accounts []PullAccountBalanceWorkItem

	// Query the database for all accounts with bank accounts that have a link type of plaid.
	_, err := j.db.Query(&accounts, `
		SELECT
			"links"."account_id",
			array_agg("links"."link_id") "link_ids"
		FROM "links"
		WHERE "links"."link_type" = ? AND "links"."plaid_link_id" IS NOT NULL
		GROUP BY "links"."account_id"
	`, models.PlaidLinkType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve accounts to update balances")
	}

	return accounts, nil
}

func (j *jobManagerBase) enqueuePullAccountBalances(job *work.Job) error {
	log := j.getLogForJob(job)

	accounts, err := j.getPlaidLinksByAccount()
	if err != nil {
		log.WithError(err).Errorf("failed to retrieve bank accounts that need to by synced")
		return err
	}

	log.Infof("enqueueing %d account(s) for sync", len(accounts))

	for _, account := range accounts {
		for _, linkId := range account.LinkIDs {
			accountLog := log.WithFields(logrus.Fields{
				"accountId": account.AccountID,
				"linkId":    linkId,
			})
			accountLog.Trace("enqueueing for account balance update")

			_, err = j.enqueueUniqueJob(PullAccountBalances, map[string]interface{}{
				"accountId": account.AccountID,
				"linkId":    linkId,
			})
			if err != nil {
				accountLog.WithError(err).Error("could not enqueue account, data will not be synced")
				continue
			}
			accountLog.Trace("successfully enqueued account for account balance update")
		}
	}

	return nil
}

func (j *jobManagerBase) pullAccountBalances(job *work.Job) error {
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Pull Account Balances"))
	defer span.Finish()

	start := time.Now()
	log := j.getLogForJob(job)
	log.Infof("pulling account balances")

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	linkId := uint64(job.ArgInt64("linkId"))
	span.SetTag("linkId", strconv.FormatUint(linkId, 10))
	span.SetTag("accountId", strconv.FormatUint(accountId, 10))

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(span.Context(), linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve link details to pull balances")
			return err
		}

		log = log.WithField("linkId", link.LinkId)

		if link.PlaidLink == nil {
			err = errors.Errorf("cannot pull account balanaces for link without plaid info")
			log.WithError(err).Errorf("failed to pull balances")
			return err
		}

		accessToken, err := j.plaidSecrets.GetAccessTokenForPlaidLinkId(span.Context(), accountId, link.PlaidLink.ItemId)
		if err != nil {
			log.WithError(err).Errorf("failed to retrieve access token for link")
			return err
		}

		bankAccounts, err := repo.GetBankAccountsByLinkId(linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve bank account details to pull balances")
			return err
		}

		// Gather the plaid account Ids so we can precisely query plaid.
		plaidIdsToBank := map[string]models.BankAccount{}
		itemBankAccountIds := make([]string, len(bankAccounts))
		for i, bankAccount := range bankAccounts {
			itemBankAccountIds[i] = bankAccount.PlaidAccountId
			plaidIdsToBank[bankAccount.PlaidAccountId] = bankAccount
		}

		log.Debugf("requesting information for %d bank account(s)", len(itemBankAccountIds))

		result, err := j.plaidClient.GetAccounts(
			span.Context(),
			accessToken,
			plaid.GetAccountsOptions{
				AccountIDs: itemBankAccountIds,
			},
		)
		if err != nil {
			log.WithError(err).Error("failed to retrieve bank accounts from plaid")
			switch plaidErr := errors.Cause(err).(type) {
			case plaid.Error:
				switch plaidErr.ErrorType {
				case "ITEM_ERROR":
					link.LinkStatus = models.LinkStatusError
					link.ErrorCode = &plaidErr.ErrorCode
					if updateErr := repo.UpdateLink(link); updateErr != nil {
						log.WithError(updateErr).Error("failed to update link to be an error state")
					}
				}
			}

			return errors.Wrap(err, "failed to retrieve bank accounts from plaid")
		}

		updatedBankAccounts := make([]models.BankAccount, 0, len(result))
		for _, item := range result {
			bankAccount := plaidIdsToBank[item.AccountID]
			bankLog := log.WithFields(logrus.Fields{
				"bankAccountId": bankAccount.BankAccountId,
				"linkId":        bankAccount.LinkId,
			})
			shouldUpdate := false
			available := int64(item.Balances.Available * 100)
			current := int64(item.Balances.Current * 100)

			if bankAccount.CurrentBalance != current {
				bankLog = bankLog.WithField("currentBalanceChanged", true)
				shouldUpdate = true
			} else {
				bankLog = bankLog.WithField("currentBalanceChanged", false)
			}

			if bankAccount.AvailableBalance != available {
				bankLog = bankLog.WithField("availableBalanceChanged", true)
				shouldUpdate = true
			} else {
				bankLog = bankLog.WithField("availableBalanceChanged", false)
			}

			bankLog = bankLog.WithField("willUpdate", shouldUpdate)

			if shouldUpdate {
				bankLog.Info("updating bank account balances")
			} else {
				bankLog.Trace("balances do not need to be updated")
			}

			if shouldUpdate {
				updatedBankAccounts = append(updatedBankAccounts, models.BankAccount{
					BankAccountId:    bankAccount.BankAccountId,
					AccountId:        accountId,
					AvailableBalance: available,
					CurrentBalance:   current,
					LastUpdated:      start.UTC(),
				})
			}
		}

		if err := repo.UpdateBankAccounts(updatedBankAccounts); err != nil {
			log.WithError(err).Error("failed to update bank account balances")
			return err
		}

		return nil
	})
}
