package jobs

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetr/rest-api/pkg/crumbs"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/monetr/rest-api/pkg/models"
	"github.com/monetr/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	PullHistoricalTransactions = "PullHistoricalTransactions"
)

func (j *jobManagerBase) TriggerPullHistoricalTransactions(accountId, linkId uint64) (jobId string, err error) {
	log := j.log.WithFields(logrus.Fields{
		"accountId": accountId,
		"linkId":    linkId,
	})

	log.Infof("queueing pull historical transactions for account")
	job, err := j.queue.EnqueueUnique(PullHistoricalTransactions, map[string]interface{}{
		"accountId": accountId,
		"linkId":    linkId,
	})
	if err != nil {
		log.WithError(err).Error("failed to enqueue pulling historical transactions")
		return "", errors.Wrap(err, "failed to enqueue pulling historical transactions")
	}

	return job.ID, nil
}

func (j *jobManagerBase) pullHistoricalTransactions(job *work.Job) (err error) {
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Pull Historical Transactions"))
	defer span.Finish()

	defer func() {
		if err != nil {
			hub.CaptureException(err)
		}
	}()

	log := j.getLogForJob(job)
	log.Infof("pulling historical transactions")

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

	twoYearsAgo := time.Now().Add(-2 * 365 * 24 * time.Hour).UTC()

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(span.Context(), linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve link details to pull historical transactions")
			return err
		}

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


		platypus, err := j.plaidClient.NewClient(span.Context(), link, accessToken)
		if err != nil {
			log.WithError(err).Error("failed to create plaid client for link")
			return err
		}

		transactions, err := platypus.GetAllTransactions(span.Context(), twoYearsAgo, time.Now(), itemBankAccountIds)
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
