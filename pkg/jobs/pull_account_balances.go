package jobs

import (
	"github.com/sirupsen/logrus"

	"github.com/gocraft/work"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/models"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/repository"
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
			"links"."account_id"
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
	log := j.getLogForJob(job)
	log.Infof("pulling account balances")

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	linkId := uint64(job.ArgInt64("linkId"))

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve link details to pull balances")
			return err
		}

		if link.PlaidLink == nil {
			err = errors.Errorf("cannot pull account balanaces for link without plaid info")
			log.WithError(err).Errorf("failed to pull balances")
			return err
		}

		bankAccounts, err := repo.GetBankAccountsByLinkId(linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve bank account details to pull balances")
			return err
		}

		groupedByPlaidAccessToken := map[string][]models.BankAccount{}

		for _, bankAccount := range bankAccounts {
			if bankAccount.Link == nil || bankAccount.Link.PlaidLink == nil {
				// TODO (elliotcourant) Log something here maybe? This shouldn't happen so we might want to try to keep track
				//  of it if it does?
				continue
			}

			accounts, ok := groupedByPlaidAccessToken[bankAccount.Link.PlaidLink.AccessToken]
			if !ok {
				// If the access token is not present, store it and create a new array.
				groupedByPlaidAccessToken[bankAccount.Link.PlaidLink.AccessToken] = []models.BankAccount{
					bankAccount,
				}

				// Keep moving along.
				continue
			}

			// If the access token is already present, simply append this account.
			accounts = append(accounts, bankAccount)
		}

		for accessToken, banks := range groupedByPlaidAccessToken {
			// Gather the plaid account Ids so we can precisely query plaid.
			plaidIdsToBankIds := map[string]uint64{}
			itemBankAccountIds := make([]string, len(banks))
			for i, bankAccount := range banks {
				itemBankAccountIds[i] = bankAccount.PlaidAccountId
				plaidIdsToBankIds[bankAccount.PlaidAccountId] = bankAccount.BankAccountId
			}

			log.Tracef("requesting information for %d bank account(s)", len(itemBankAccountIds))

			result, err := j.plaidClient.GetAccountsWithOptions(
				accessToken,
				plaid.GetAccountsOptions{
					AccountIDs: itemBankAccountIds,
				},
			)
			if err != nil {
				log.WithError(err).Error("failed to retrieve bank accounts from plaid")
				continue // We don't want it to prevent others from being processed.
			}

			// TODO (elliotcourant) If we lift this array out of this loop then we could update all the bank accounts for all
			//  of an account's links in a single query which would be more efficient.
			updatedBankAccounts := make([]models.BankAccount, len(result.Accounts))
			for i, item := range result.Accounts {
				// TODO (elliotcourant) Maybe add something here to compare balances to the existing account record? If there
				//  are no changes there is no need to update the account at all.
				updatedBankAccounts[i] = models.BankAccount{
					BankAccountId:    plaidIdsToBankIds[item.AccountID],
					AccountId:        accountId,
					AvailableBalance: int64(item.Balances.Available * 100),
					CurrentBalance:   int64(item.Balances.Current * 100),
				}
			}

			if err := repo.UpdateBankAccounts(updatedBankAccounts); err != nil {
				log.WithError(err).Error("failed to update bank account balances")
				return err
			}
		}

		return nil
	})
}
