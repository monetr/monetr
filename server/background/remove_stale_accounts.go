package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/repository"
	"github.com/sirupsen/logrus"
)

var (
	_ ScheduledJobHandler = &RemoveStaleAccountsHandler{}
	_ JobImplementation   = &RemoveStaleAccountsJob{}
)

const (
	RemoveStaleAccounts = "RemoveStaleAccounts"
)

type (
	RemoveStaleAccountsHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		configuration config.Configuration
		enqueuer      JobEnqueuer
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	RemoveStaleAccountsJob struct {
		log      *logrus.Entry
		repo     repository.JobRepository
		enqueuer JobEnqueuer
		clock    clock.Clock
	}
)

func NewRemoveStaleAccountsHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	configuration config.Configuration,
	enqueuer JobEnqueuer,
) *RemoveStaleAccountsHandler {
	return &RemoveStaleAccountsHandler{
		log:           log,
		db:            db,
		configuration: configuration,
		enqueuer:      enqueuer,
		unmarshaller:  DefaultJobUnmarshaller,
		clock:         clock,
	}
}

// DefaultSchedule implements ScheduledJobHandler.
func (r *RemoveStaleAccountsHandler) DefaultSchedule() string {
	// Every day at 12:45AM
	return "0 45 0 * * *"
}

// EnqueueTriggeredJob implements ScheduledJobHandler.
func (r *RemoveStaleAccountsHandler) EnqueueTriggeredJob(
	ctx context.Context,
	enqueuer JobEnqueuer,
) error {
	return enqueuer.EnqueueJob(ctx, r.QueueName(), nil)
}

// HandleConsumeJob implements ScheduledJobHandler.
func (r *RemoveStaleAccountsHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	span := sentry.StartSpan(ctx, "db.transaction")
	defer span.Finish()

	job := NewRemoveStaleAccountsJob(
		log.WithContext(span.Context()),
		r.db,
		r.clock,
		r.enqueuer,
	)
	return job.Run(span.Context())
}

// QueueName implements ScheduledJobHandler.
func (r *RemoveStaleAccountsHandler) QueueName() string {
	return RemoveStaleAccounts
}

func NewRemoveStaleAccountsJob(
	log *logrus.Entry,
	db pg.DBI,
	clock clock.Clock,
	enqueuer JobEnqueuer,
) *RemoveStaleAccountsJob {
	return &RemoveStaleAccountsJob{
		log:      log,
		repo:     repository.NewJobRepository(db, clock),
		enqueuer: enqueuer,
		clock:    clock,
	}
}

// Run implements JobImplementation.
func (r *RemoveStaleAccountsJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := r.log.WithContext(span.Context())
	log.Info("looking for stale accounts that need to be cleaned up")

	staleAccounts, err := r.repo.GetStaleAccounts(span.Context())
	if err != nil {
		return err
	}

	if len(staleAccounts) == 0 {
		log.Info("no stale accounts to be cleaned up at this time!")
		return nil
	}

	log.WithField("count", len(staleAccounts)).Info("found stale accounts to be cleaned up!")

	for _, item := range staleAccounts {
		itemLog := log.WithFields(logrus.Fields{
			"accountId": item.AccountId,
			"account": logrus.Fields{
				"trialEndsAt":             item.TrialEndsAt,
				"subscriptionActiveUntil": item.SubscriptionActiveUntil,
			},
		})
		itemLog.Debug("enqueuing stale account for removal")
		err := r.enqueuer.EnqueueJob(ctx, RemoveAccount, RemoveAccountArguments{
			AccountId: item.AccountId,
		})
		if err != nil {
			itemLog.WithError(err).Warn("failed to enqueue job to remove account")
			crumbs.Warn(ctx, "Failed to enqueue job to remove account", "job", map[string]any{
				"error":     err,
				"accountId": item.AccountId,
				"account": logrus.Fields{
					"trialEndsAt":             item.TrialEndsAt,
					"subscriptionActiveUntil": item.SubscriptionActiveUntil,
				},
			})
			continue
		}
		itemLog.Trace("successfully enqueued job for account removal")
	}

	return nil
}
