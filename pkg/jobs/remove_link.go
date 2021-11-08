package jobs

import (
	"context"
	"fmt"
	"strconv"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/gocraft/work"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/models"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	RemoveLink = "RemoveLink"
)

func (j *jobManagerBase) TriggerRemoveLink(accountId, userId, linkId uint64) (jobId string, err error) {
	job, err := j.queue.EnqueueUnique(RemoveLink, map[string]interface{}{
		"accountId": accountId,
		"userId":    userId,
		"linkId":    linkId,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to enqueue link removal")
	}

	return job.ID, nil
}

type RemoveLinkJob struct {
	jobId     string
	accountId uint64
	linkId    uint64
	userId    uint64
	log       *logrus.Entry
	db        *pg.DB
	notify    pubsub.Publisher
}

func (r *RemoveLinkJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "Job", sentry.TransactionName("Remove Link"))
	defer span.Finish()

	span.SetTag("jobId", r.jobId)
	span.SetTag("linkId", strconv.FormatUint(r.linkId, 10))
	span.SetTag("accountId", strconv.FormatUint(r.accountId, 10))

	if hub := sentry.GetHubFromContext(span.Context()); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       strconv.FormatUint(r.accountId, 10),
				Username: fmt.Sprintf("account:%d", r.accountId),
			})
		})
	}

	log := r.log

	return r.db.RunInTransaction(span.Context(), func(txn *pg.Tx) error {
		repo := repository.NewRepositoryFromSession(r.userId, r.accountId, txn)

		link, err := repo.GetLink(span.Context(), r.linkId)
		if err != nil {
			crumbs.Warn(span.Context(), "failed to retrieve link to be removed, this job will not be retried", "weirdness", nil)
			log.WithError(err).Error("failed to retrieve link that to be removed, this job will not be retried")
			return nil
		}

		if link.PlaidLink != nil {
			crumbs.IncludePlaidItemIDTag(span, link.PlaidLink.ItemId)
		}

		bankAccountIds := make([]uint64, 0)
		{
			err = txn.ModelContext(span.Context(), &models.BankAccount{}).
				Where(`"bank_account"."account_id" = ?`, r.accountId).
				Where(`"bank_account"."link_id" = ?`, r.linkId).
				Column("bank_account_id").
				Select(&bankAccountIds)
			if err != nil {
				log.WithError(err).Errorf("failed to retrieve bank account Ids for link")
				return errors.Wrap(err, "failed to retrieve bank account Ids for link")
			}
		}

		if len(bankAccountIds) > 0 {
			log.WithField("bankAccountIds", bankAccountIds).Info("removing data for bank account Ids for link")

			{
				result, err := txn.ModelContext(span.Context(), &models.Transaction{}).
					Where(`"transaction"."account_id" = ?`, r.accountId).
					WhereIn(`"transaction"."bank_account_id" IN (?)`, bankAccountIds).
					Delete()
				if err != nil {
					log.WithError(err).Errorf("failed to remove transactions for link")
					return errors.Wrap(err, "failed to remove transactions for link")
				}

				log.WithField("removed", result.RowsAffected()).Info("removed transaction(s)")
			}

			{
				result, err := txn.ModelContext(span.Context(), &models.Spending{}).
					Where(`"spending"."account_id" = ?`, r.accountId).
					WhereIn(`"spending"."bank_account_id" IN (?)`, bankAccountIds).
					Delete()
				if err != nil {
					log.WithError(err).Errorf("failed to remove spending for link")
					return errors.Wrap(err, "failed to remove spending for link")
				}

				log.WithField("removed", result.RowsAffected()).Info("removed spending(s)")
			}

			{
				result, err := txn.ModelContext(span.Context(), &models.FundingSchedule{}).
					Where(`"funding_schedule"."account_id" = ?`, r.accountId).
					WhereIn(`"funding_schedule"."bank_account_id" IN (?)`, bankAccountIds).
					Delete()
				if err != nil {
					log.WithError(err).Errorf("failed to remove funding schedules for link")
					return errors.Wrap(err, "failed to remove funding schedules for link")
				}

				log.WithField("removed", result.RowsAffected()).Info("removed funding schedule(s)")
			}

			{
				result, err := txn.ModelContext(span.Context(), &models.BankAccount{}).
					Where(`"bank_account"."account_id" = ?`, r.accountId).
					WhereIn(`"bank_account"."bank_account_id" IN (?)`, bankAccountIds).
					Delete()
				if err != nil {
					log.WithError(err).Errorf("failed to remove bank accounts for link")
					return errors.Wrap(err, "failed to remove bank accounts for link")
				}

				log.WithField("removed", result.RowsAffected()).Info("removed bank account(s)")
			}
		} else {
			crumbs.Debug(span.Context(), "There were no bank accounts associated with this link.", map[string]interface{}{})
			log.Info("no bank accounts associated with link, deleting link")
		}

		{
			// Delete the link directly, I don't want to include something like this on the repository interface as it is not
			// something that I want just anything to be able to do. Deleting a link has the potential to move a ton of data
			// and should be done in the background. This should do a cascade if I setup my foreign keys correctly but tests
			// should be written to verify that those cascades are _always_ happening properly.
			result, err := txn.ModelContext(span.Context(), link).
				WherePK().
				Delete(&link)
			if err != nil {
				log.WithError(err).Error("failed to delete link")
				return err
			}
			log.WithField("removed", result.RowsAffected()).Info("successfully removed link")
		}

		channelName := fmt.Sprintf("link:remove:%d:%d", r.accountId, r.linkId)

		if err = r.notify.Notify(span.Context(), channelName, "success"); err != nil {
			log.WithError(err).Warn("failed to send notification about successfully removing link")
			crumbs.Warn(span.Context(), "failed to send notification about successfully removing link", "pubsub", map[string]interface{}{
				"error": err.Error(),
			})
		}

		return nil
	})
}

func (j *jobManagerBase) newRemoveLinkJob(job *work.Job) (*RemoveLinkJob, error) {
	log := j.getLogForJob(job)

	// We need to know what link we are actually deleting. This is required.
	linkId := uint64(job.ArgInt64("linkId"))
	if linkId == 0 {
		log.Error("link Id is 0, the link Id must be specified for removal")
		return nil, errors.New("must specify link Id for removal")
	}

	// Deleting links is never an automatic process. This is always initiated by a user, so we want to include the user
	// who initiatied this action in our logs so we have some record of _who_ initiated the delete.
	userId := uint64(job.ArgInt64("userId"))
	if userId == 0 {
		log.Error("user Id is 0, the user Id must be specified for removal")
		return nil, errors.New("must specify user Id for removal")
	}

	// Validate that we also have an accountId.
	accountId, err := j.getAccountId(job)
	if err != nil {
		log.WithError(err).Error("could not run job, no account Id")
		return nil, err
	}

	log = log.WithFields(logrus.Fields{
		"accountId": accountId,
		"linkId":    linkId,
		"userId":    userId,
	})

	runner := &RemoveLinkJob{
		jobId:     job.ID,
		accountId: accountId,
		linkId:    linkId,
		userId:    userId,
		log:       log,
		db:        j.db,
		notify:    j.ps,
	}

	return runner, nil
}

func (j *jobManagerBase) removeLink(input *work.Job) error {
	job, err := j.newRemoveLinkJob(input)
	if err != nil {
		return err
	}

	return job.Run(context.Background())
}
