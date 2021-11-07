package jobs

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
)

const (
	PullInitialTransactions = "PullInitialTransactions"
)

func (j *jobManagerBase) pullInitialTransactions(job *work.Job) (err error) {
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Pull Initial Transactions"))
	defer span.Finish()

	defer func() {
		if err != nil {
			hub.CaptureException(err)
		}
	}()

	log := j.getLogForJob(job)

	accountId := uint64(job.ArgInt64("accountId"))

	linkId := uint64(job.ArgInt64("linkId"))
	if linkId == 0 {
		log.Error("cannot pull initial transactions without a link Id")
		return errors.Errorf("cannot pull initial transactions without a link Id")
	}

	log = log.WithField("linkId", linkId)

	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       strconv.FormatUint(accountId, 10),
			Username: fmt.Sprintf("account:%d", accountId),
		})
		scope.SetTag("accountId", strconv.FormatUint(accountId, 10))
		scope.SetTag("linkId", strconv.FormatUint(linkId, 10))
		scope.SetTag("jobId", job.ID)
	})

	err = j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(span.Context(), linkId)
		if err != nil {
			log.WithError(err).Error("cannot pull initial transactions for link provided")
			return nil
		}

		if link.LinkStatus == models.LinkStatusSetup {
			log.Warn("link has already been setup, initial transactions will not be pulled")
			return nil
		}

		if link.PlaidLink == nil {
			log.Error("provided link does not have any plaid credentials")
			return nil
		}

		accessToken, err := j.plaidSecrets.GetAccessTokenForPlaidLinkId(span.Context(), accountId, link.PlaidLink.ItemId)
		if err != nil {
			log.WithError(err).Errorf("failed to retrieve access token for link")
			return err
		}

		if len(link.BankAccounts) == 0 {
			log.Error("no bank accounts for plaid link")
			return nil
		}

		plaidIdsToBankIds := map[string]uint64{}
		bankAccountIds := make([]string, len(link.BankAccounts))
		for i, bankAccount := range link.BankAccounts {
			bankAccountIds[i] = bankAccount.PlaidAccountId
			plaidIdsToBankIds[bankAccount.PlaidAccountId] = bankAccount.BankAccountId
		}

		now := time.Now().UTC()
		platypus, err := j.plaidClient.NewClient(span.Context(), link, accessToken, link.PlaidLink.ItemId)
		if err != nil {
			log.WithError(err).Error("failed to create plaid client for link")
			return err
		}

		plaidTransactions, err := platypus.GetAllTransactions(span.Context(), now.Add(-30*24*time.Hour), now, bankAccountIds)
		if err != nil {
			log.WithError(err).Error("failed to retrieve initial transactions from plaid")
			return errors.Wrap(err, "failed to retrieve initial transactions from plaid")
		}

		if len(plaidTransactions) == 0 {
			log.Warn("no transactions were retrieved from plaid")
			return nil
		}

		log.Debugf("retreived %d transaction(s) from plaid, processing now", len(plaidTransactions))

		if err = j.upsertTransactions(
			span.Context(),
			log,
			repo,
			link,
			plaidIdsToBankIds,
			plaidTransactions,
		); err != nil {
			log.WithError(err).Error("failed to upsert transactions from plaid")
			return err
		}

		link.LinkStatus = models.LinkStatusSetup
		link.LastSuccessfulUpdate = myownsanity.TimeP(time.Now().UTC())
		if err = repo.UpdateLink(span.Context(), link); err != nil {
			log.WithError(err).Error("failed to update link status")
			return err
		}

		return nil
	})

	{ // Send the notification that the link has been set up.
		time.Sleep(5 * time.Second)

		channelName := fmt.Sprintf("initial:plaid:link:%d:%d", accountId, linkId)
		if notifyErr := j.ps.Notify(
			span.Context(),
			channelName,
			"success",
		); notifyErr != nil {
			log.WithError(notifyErr).Error("failed to publish link status to pubsub")
		}
	}

	return err
}
