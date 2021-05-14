package jobs

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/monetrapp/rest-api/pkg/repository"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

const (
	RemoveTransactions = "RemoveTransactions"
)

func (j *jobManagerBase) TriggerRemoveTransactions(accountId, linkId uint64, removedTransactions []string) (jobId string, err error) {
	job, err := j.queue.EnqueueUnique(RemoveTransactions, map[string]interface{}{
		"accountId":           accountId,
		"linkId":              linkId,
		"removedTransactions": strings.Join(removedTransactions, ","),
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to enqueue transaction removal")
	}

	return job.ID, nil
}

func (j *jobManagerBase) removeTransactions(job *work.Job) error {
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(context.Background(), hub)
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Remove Transactions"))
	defer span.Finish()

	start := time.Now()
	log := j.getLogForJob(job)

	transactionIds := strings.Split(job.ArgString("removedTransactions"), ",")

	log.Infof("removing %d transaction(s)", len(transactionIds))

	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return err
	}

	defer func() {
		if j.stats != nil {
			j.stats.JobFinished(RemoveTransactions, accountId, start)
		}
	}()

	linkId := uint64(job.ArgInt64("linkId"))
	span.SetTag("accountId", strconv.FormatUint(accountId, 10))
	span.SetTag("linkId", strconv.FormatUint(linkId, 10))

	return j.getRepositoryForJob(job, func(repo repository.Repository) error {
		link, err := repo.GetLink(span.Context(), linkId)
		if err != nil {
			log.WithError(err).Error("failed to retrieve link details to pull transactions")
			return err
		}

		if link.PlaidLink == nil {
			err = errors.Errorf("cannot pull account balanaces for link without plaid info")
			log.WithError(err).Errorf("failed to pull transactions")
			return err
		}

		transactions, err := repo.GetTransactionsByPlaidTransactionId(span.Context(), linkId, transactionIds)
		if err != nil {
			log.WithError(err).Error("failed to retrieve transactions by plaid transaction Id for removal")
			return err
		}

		if len(transactions) == 0 {
			log.Warnf("no transactions retrieved, nothing to be done. transactions might already have been deleted")
			return nil
		}

		if len(transactions) != len(transactionIds) {
			log.Warnf("number of transactions retrieved does not match expected number of transactions, expected: %d found: %d", len(transactionIds), len(transactions))
		}

		for _, existingTransaction := range transactions {
			if existingTransaction.SpendingId == nil {
				continue
			}

			// If the transaction is spent from something then we need to remove the spent from before deleting it to
			// maintain our balances correctly.
			updatedTransaction := existingTransaction
			updatedTransaction.SpendingId = nil

			// This is a simple sanity check, working with objects in slices and for loops can be goofy, or my
			// understanding of the way objects works with how they are referenced in memory is poor. This is to make
			// sure im not doing it wrong though. I'm worried that making a "copy" of the object and then modifying the
			// copy will modify the original as well.
			if existingTransaction.SpendingId == nil {
				sentry.CaptureMessage("original transaction modified")
				panic("original transaction modified")
			}

			_, err = repo.ProcessTransactionSpentFrom(
				span.Context(),
				existingTransaction.BankAccountId,
				&updatedTransaction,
				&existingTransaction,
			)
			if err != nil {
				return err
			}
		}

		for _, transaction := range transactions {
			if err := repo.DeleteTransaction(span.Context(), transaction.BankAccountId, transaction.TransactionId); err != nil {
				log.WithField("transactionId", transaction.TransactionId).WithError(err).
					Error("failed to delete transaction")
				return err
			}
		}

		log.Debugf("successfully removed %d transaction(s)", len(transactions))

		return nil
	})
}
