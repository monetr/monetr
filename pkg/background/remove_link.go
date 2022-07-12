package background

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
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

var (
	_ JobHandler = &RemoveLinkHandler{}
	_ Job        = &RemoveLinkJob{}
)

type (
	RemoveLinkHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		publisher    pubsub.Publisher
		unmarshaller JobUnmarshaller
	}

	RemoveLinkArguments struct {
		AccountId uint64 `json:"accountId"`
		LinkId    uint64 `json:"linkId"`
	}

	RemoveLinkJob struct {
		args      RemoveLinkArguments
		log       *logrus.Entry
		db        pg.DBI
		publisher pubsub.Publisher
	}
)

// TriggerRemoveLink will dispatch a background job to remove the specified link and all of the data related to it from
// the desired account. This will return an error if the job fails to be enqueued, but does not indicate the status of
// the actual job.
func TriggerRemoveLink(ctx context.Context, backgroundJobs JobController, arguments RemoveLinkArguments) error {
	return backgroundJobs.triggerJob(ctx, RemoveLink, arguments)
}

func NewRemoveLinkHandler(
	log *logrus.Entry,
	db *pg.DB,
	publisher pubsub.Publisher,
) *RemoveLinkHandler {
	return &RemoveLinkHandler{
		log:          log,
		db:           db,
		publisher:    publisher,
		unmarshaller: DefaultJobUnmarshaller,
	}
}

func (r *RemoveLinkHandler) SetUnmarshaller(unmarshaller JobUnmarshaller) {
	r.unmarshaller = unmarshaller
}

func (r RemoveLinkHandler) QueueName() string {
	return RemoveLink
}

func (r *RemoveLinkHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	var args RemoveLinkArguments
	if err := errors.Wrap(r.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Remove Link job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return r.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		job, err := NewRemoveLinkJob(r.log.WithContext(span.Context()), txn, r.publisher, args)
		if err != nil {
			return err
		}

		return job.Run(span.Context())
	})
}

func NewRemoveLinkJob(
	log *logrus.Entry,
	db pg.DBI,
	publisher pubsub.Publisher,
	args RemoveLinkArguments,
) (*RemoveLinkJob, error) {
	return &RemoveLinkJob{
		args:      args,
		log:       log,
		db:        db,
		publisher: publisher,
	}, nil
}

func (r *RemoveLinkJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	accountId := r.args.AccountId
	linkId := r.args.LinkId

	repo := repository.NewRepositoryFromSession(0, accountId, r.db)

	log := r.log.WithContext(span.Context())

	link, err := repo.GetLink(span.Context(), linkId)
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
		err = r.db.ModelContext(span.Context(), &models.BankAccount{}).
			Where(`"bank_account"."account_id" = ?`, accountId).
			Where(`"bank_account"."link_id" = ?`, linkId).
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
			result, err := r.db.ModelContext(span.Context(), &models.Transaction{}).
				Where(`"transaction"."account_id" = ?`, accountId).
				WhereIn(`"transaction"."bank_account_id" IN (?)`, bankAccountIds).
				Delete()
			if err != nil {
				log.WithError(err).Errorf("failed to remove transactions for link")
				return errors.Wrap(err, "failed to remove transactions for link")
			}

			log.WithField("removed", result.RowsAffected()).Info("removed transaction(s)")
		}

		{
			result, err := r.db.ModelContext(span.Context(), &models.Spending{}).
				Where(`"spending"."account_id" = ?`, accountId).
				WhereIn(`"spending"."bank_account_id" IN (?)`, bankAccountIds).
				Delete()
			if err != nil {
				log.WithError(err).Errorf("failed to remove spending for link")
				return errors.Wrap(err, "failed to remove spending for link")
			}

			log.WithField("removed", result.RowsAffected()).Info("removed spending(s)")
		}

		{
			result, err := r.db.ModelContext(span.Context(), &models.FundingSchedule{}).
				Where(`"funding_schedule"."account_id" = ?`, accountId).
				WhereIn(`"funding_schedule"."bank_account_id" IN (?)`, bankAccountIds).
				Delete()
			if err != nil {
				log.WithError(err).Errorf("failed to remove funding schedules for link")
				return errors.Wrap(err, "failed to remove funding schedules for link")
			}

			log.WithField("removed", result.RowsAffected()).Info("removed funding schedule(s)")
		}

		{
			result, err := r.db.ModelContext(span.Context(), &models.BankAccount{}).
				Where(`"bank_account"."account_id" = ?`, accountId).
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
		result, err := r.db.ModelContext(span.Context(), link).
			WherePK().
			Delete(&link)
		if err != nil {
			log.WithError(err).Error("failed to delete link")
			return err
		}
		log.WithField("removed", result.RowsAffected()).Info("successfully removed link")
	}

	channelName := fmt.Sprintf("link:remove:%d:%d", accountId, linkId)
	if err = r.publisher.Notify(span.Context(), channelName, "success"); err != nil {
		log.WithError(err).Warn("failed to send notification about successfully removing link")
		crumbs.Warn(span.Context(), "failed to send notification about successfully removing link", "pubsub", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return nil
}
