package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// CleanupLunchFlow is a job that runs nightly and just removes any Lunch Flow
	// links that are in a pending status. These serve no purpose and are
	// persisted only as part of the setup process. If they have not been moved to
	// active after so much time then they can be removed safely.
	CleanupLunchFlow = "CleanupLunchFlow"
)

var (
	_ ScheduledJobHandler = &CleanupLunchFlowHandler{}
	_ JobImplementation   = &CleanupLunchFlowJob{}
)

type (
	CleanupLunchFlowHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		kms          secrets.KeyManagement
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	CleanupLunchFlowArguments struct {
		AccountId       ID[Account]       `json:"accountId"`
		LunchFlowLinkId ID[LunchFlowLink] `json:"lunchFlowLinkId"`
	}

	CleanupLunchFlowJob struct {
		args    CleanupLunchFlowArguments
		log     *logrus.Entry
		clock   clock.Clock
		repo    repository.BaseRepository
		secrets repository.SecretsRepository
	}
)

func NewCleanupLunchFlowHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	kms secrets.KeyManagement,
) *CleanupLunchFlowHandler {
	return &CleanupLunchFlowHandler{
		log:          log,
		db:           db,
		kms:          kms,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

// DefaultSchedule implements ScheduledJobHandler.
func (c *CleanupLunchFlowHandler) DefaultSchedule() string {
	// Run every day at 1:15AM
	return "0 15 1 * * *"
}

// EnqueueTriggeredJob implements ScheduledJobHandler.
func (c *CleanupLunchFlowHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	log := c.log.WithContext(ctx)

	log.Info("retrieving bank accounts to sync with Lunch Flow")

	jobRepo := repository.NewJobRepository(c.db, c.clock)

	staleLinks, err := jobRepo.GetStaleLunchFlowLinks(ctx)
	if err != nil {
		return err
	}

	if len(staleLinks) == 0 {
		log.Info("no stale Lunch Flow links to be cleaned up")
		return nil
	}

	for _, item := range staleLinks {
		itemLog := log.WithFields(logrus.Fields{
			"accountId":       item.AccountId,
			"lunchFlowLinkId": item.LunchFlowLinkId,
		})

		itemLog.Trace("enqueuing Lunch Flow link to be cleand up")

		err := enqueuer.EnqueueJobTxn(
			ctx,
			c.db,
			c.QueueName(),
			CleanupLunchFlowArguments{
				AccountId:       item.AccountId,
				LunchFlowLinkId: item.LunchFlowLinkId,
			},
		)
		if err != nil {
			itemLog.WithError(err).Warn("failed to enqueue job to cleanup Lunch Flow link")
			crumbs.Warn(
				ctx,
				"Failed to enqueue job to cleanup Lunch Flow link",
				"job",
				map[string]any{
					"error": err,
				},
			)
			continue
		}

		itemLog.Trace("successfully enqueued Lunch Flow link for cleanup")
	}

	return nil
}

// HandleConsumeJob implements ScheduledJobHandler.
func (c *CleanupLunchFlowHandler) HandleConsumeJob(ctx context.Context, log *logrus.Entry, data []byte) error {
	var args CleanupLunchFlowArguments
	if err := errors.Wrap(c.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for cleanup Lunch Flow link job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)
	log = log.WithFields(logrus.Fields{
		"accountId":       args.AccountId,
		"lunchFlowLinkId": args.LunchFlowLinkId,
	})

	return c.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.WithContext(span.Context())
		repo := repository.NewRepositoryFromSession(
			c.clock,
			"user_lunch_flow",
			args.AccountId,
			txn,
			log,
		)
		secretsRepo := repository.NewSecretsRepository(
			log,
			c.clock,
			txn,
			c.kms,
			args.AccountId,
		)

		job, err := NewCleanupLunchFlowJob(
			log,
			repo,
			secretsRepo,
			c.clock,
			args,
		)
		if err != nil {
			return err
		}

		return job.Run(span.Context())
	})
}

// QueueName implements ScheduledJobHandler.
func (c *CleanupLunchFlowHandler) QueueName() string {
	return CleanupLunchFlow
}

func NewCleanupLunchFlowJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	secrets repository.SecretsRepository,
	clock clock.Clock,
	args CleanupLunchFlowArguments,
) (*CleanupLunchFlowJob, error) {
	return &CleanupLunchFlowJob{
		args:  args,
		log:   log,
		clock: clock,
		repo:  repo,
	}, nil
}

// Run implements JobImplementation.
func (c *CleanupLunchFlowJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	lunchFlowLink, err := c.repo.GetLunchFlowLink(
		span.Context(),
		c.args.LunchFlowLinkId,
	)
	if err != nil {
		return err
	}

	log := c.log.WithContext(span.Context()).
		WithFields(logrus.Fields{
			"lunchFlowLinkId": lunchFlowLink.LunchFlowLinkId,
		})

	if lunchFlowLink.Status != LunchFlowLinkStatusPending {
		log.WithField("bug", true).Warn("Lunch Flow link is not in a pending status! There is a bug, please report this!")
		return nil
	}

	lunchFlowBankAccounts, err := c.repo.GetLunchFlowBankAccountsByLunchFlowLink(
		span.Context(),
		lunchFlowLink.LunchFlowLinkId,
	)
	if err != nil {
		return err
	}

	for _, lunchFlowBankAccount := range lunchFlowBankAccounts {
		if err := c.repo.DeleteLunchFlowBankAccount(
			span.Context(),
			lunchFlowBankAccount.LunchFlowBankAccountId,
		); err != nil {
			log.WithError(err).
				WithFields(logrus.Fields{
					"lunchFlowBankAccountId": lunchFlowBankAccount.LunchFlowBankAccountId,
				}).
				Error("failed to remove Lunch Flow bank account!")
			return err
		}
	}

	if err := c.repo.DeleteLunchFlowLink(span.Context(), lunchFlowLink.LunchFlowLinkId); err != nil {
		log.WithError(err).
			Error("failed to remove Lunch Flow link!")
		return err
	}

	if err := c.secrets.Delete(span.Context(), lunchFlowLink.SecretId); err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"secretId": lunchFlowLink.SecretId,
			}).
			Error("failed to remove Lunch Flow link secret!")
		return err
	}

	return nil
}
