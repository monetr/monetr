package jobs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/monetr/rest-api/pkg/crumbs"

	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	EnqueuePullLatestTransactions = "EnqueuePullLatestTransactions"
	PullLatestTransactions        = "PullLatestTransactions"
)

func (j *jobManagerBase) TriggerPullLatestTransactions(accountId, linkId uint64, numberOfTransactions int64) (jobId string, err error) {
	log := j.log.WithFields(logrus.Fields{
		"accountId": accountId,
		"linkId":    linkId,
	})

	log.Infof("queueing pull latest transactions for account")
	job, err := j.queue.EnqueueUnique(PullLatestTransactions, map[string]interface{}{
		"accountId":            accountId,
		"linkId":               linkId,
		"numberOfTransactions": numberOfTransactions,
	})
	if err != nil {
		log.WithError(err).Error("failed to enqueue pulling latest transactions")
		return "", errors.Wrap(err, "failed to enqueue pulling latest transactions")
	}
	log = log.WithField("pullLatestTransactionsJobId", job.ID)

	log.Infof("queueing account balances update for account")
	job, err = j.queue.EnqueueUnique(PullAccountBalances, map[string]interface{}{
		"accountId": accountId,
		"linkId":    linkId,
	})
	if err != nil {
		log.WithError(err).Error("failed to enqueue pulling account balances")
		return "", errors.Wrap(err, "failed to enqueue pulling account balances")
	}

	return job.ID, nil
}

func (j *jobManagerBase) enqueuePullLatestTransactions(job *work.Job) error {
	log := j.getLogForJob(job)

	accounts, err := j.getPlaidLinksByAccount()
	if err != nil {
		log.WithError(err).Errorf("failed to retrieve bank accounts that need to by synced")
		return err
	}

	log.Infof("enqueing %d account(s) to pull latest transactions", len(accounts))

	for _, account := range accounts {
		for _, linkId := range account.LinkIDs {
			accountLog := log.WithFields(logrus.Fields{
				"accountId": account.AccountID,
				"linkId":    linkId,
			})
			accountLog.Trace("enqueueing for latest transactions update")
			_, err := j.queue.EnqueueUnique(PullLatestTransactions, map[string]interface{}{
				"accountId": account.AccountID,
				"linkId":    linkId,
			})
			if err != nil {
				accountLog.WithError(err).Error("could not enqueue account, data will not be synced")
				continue
			}

			accountLog.Trace("successfully enqueued account for latest transactions update")
		}
	}

	return nil
}

func (j *jobManagerBase) pullLatestTransactions(job *work.Job) (err error) {
	defer func() {
		if err := recover(); err != nil {
			sentry.CaptureException(errors.Errorf("pull latest transactions failure: %+v", err))
		}
	}()

	log := j.getLogForJob(job)
	log.Infof("pulling account balances")

	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Pull Latest Transactions"))
	defer span.Finish()

	defer func() {
		if err != nil {
			hub.CaptureException(err)
		}
	}()

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	linkId := uint64(job.ArgInt64("linkId"))

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       strconv.FormatUint(accountId, 10),
			Username: fmt.Sprintf("account:%d", accountId),
		})
		scope.SetTag("accountId", strconv.FormatUint(accountId, 10))
		scope.SetTag("linkId", strconv.FormatUint(linkId, 10))
		scope.SetTag("jobId", job.ID)
	})

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(span.Context(), linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve link details to pull transactions")
			return err
		}

		log = log.WithField("linkId", link.LinkId)

		if link.PlaidLink == nil {
			err = errors.Errorf("cannot pull account balanaces for link without plaid info")
			log.WithError(err).Errorf("failed to pull transactions")
			return err
		}

		switch link.LinkStatus {
		case models.LinkStatusSetup, models.LinkStatusPendingExpiration:
			break
		default:
			crumbs.Warn(span.Context(), "Link is not in a state where data can be retrieved", "plaid", map[string]interface{}{
				"status": link.LinkStatus,
			})
			return nil
		}

		accessToken, err := j.plaidSecrets.GetAccessTokenForPlaidLinkId(span.Context(), accountId, link.PlaidLink.ItemId)
		if err != nil {
			log.WithError(err).Errorf("failed to retrieve access token for link")
			return err
		}

		bankAccounts, err := repo.GetBankAccountsByLinkId(span.Context(), linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve bank account details to pull transactions")
			return err
		}

		// Gather the plaid account Ids so we can precisely query plaid.
		plaidIdsToBankIds := map[string]uint64{}
		itemBankAccountIds := make([]string, len(bankAccounts))
		for i, bankAccount := range bankAccounts {
			itemBankAccountIds[i] = bankAccount.PlaidAccountId
			plaidIdsToBankIds[bankAccount.PlaidAccountId] = bankAccount.BankAccountId
		}

		log.Debugf("retrieving transactions for %d bank account(s)", len(itemBankAccountIds))

		// Request the last 7 days worth of transactions for update.
		start := time.Now().Add(-7 * 24 * time.Hour)
		if link.LastSuccessfulUpdate == nil {
			// But if there has not been a last successful update set yet, then request the last 30 days to handle this
			// update.
			start = time.Now().Add(-30 * 24 * time.Hour)
		} else if start.After(*link.LastSuccessfulUpdate) {
			// If we haven't seen an update in longer than 7 days, then use the last successful update date instead.
			start = *link.LastSuccessfulUpdate
		}
		end := time.Now()

		platypus, err := j.plaidClient.NewClient(span.Context(), link, accessToken)
		if err != nil {
			log.WithError(err).Error("failed to create plaid client for link")
			return err
		}

		transactions, err := platypus.GetAllTransactions(span.Context(), start, end, itemBankAccountIds)
		if err != nil {
			log.WithError(err).Error("failed to retrieve transactions from plaid")
			return errors.Wrap(err, "failed to retrieve transactions from plaid")
		}

		if err = j.upsertTransactions(
			span.Context(),
			log,
			repo,
			link,
			plaidIdsToBankIds,
			transactions,
		); err != nil {
			log.WithError(err).Error("failed to upsert transactions from plaid")
			return err
		}

		link.LastSuccessfulUpdate = myownsanity.TimeP(time.Now().UTC())
		return repo.UpdateLink(span.Context(), link)
	})
}
