package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/repository"
	"github.com/sirupsen/logrus"
)

const (
	RemoveInactiveLinks = "RemoveInactiveLinks"
)

type (
	RemoveInactiveLinksHandler struct {
		log           *logrus.Entry
		db            *pg.DB
		configuration config.Configuration
		repo          repository.JobRepository
		enqueuer      JobEnqueuer
		unmarshaller  JobUnmarshaller
		clock         clock.Clock
	}

	RemoveInactiveLinksJob struct {
		log      *logrus.Entry
		repo     repository.JobRepository
		enqueuer JobEnqueuer
		clock    clock.Clock
	}
)

func NewRemoveInactiveLinksHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	configuration config.Configuration,
	enqueuer JobEnqueuer,
) *RemoveInactiveLinksHandler {
	return &RemoveInactiveLinksHandler{
		log:           log,
		db:            db,
		configuration: configuration,
		repo:          repository.NewJobRepository(db, clock),
		enqueuer:      enqueuer,
		unmarshaller:  DefaultJobUnmarshaller,
		clock:         clock,
	}
}

func (h RemoveInactiveLinksHandler) QueueName() string {
	return RemoveInactiveLinks
}

func (h RemoveInactiveLinksHandler) DefaultSchedule() string {
	// Every day at 6:15AM
	return "0 15 6 * * *"
}

func (h RemoveInactiveLinksHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	return enqueuer.EnqueueJob(ctx, h.QueueName(), nil)
}

func (h *RemoveInactiveLinksHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	job, err := NewRemoveInactiveLinksJob(
		log,
		h.repo,
		h.clock,
		h.enqueuer,
	)
	if err != nil {
		return err
	}
	return job.Run(ctx)
}

func NewRemoveInactiveLinksJob(
	log *logrus.Entry,
	repo repository.JobRepository,
	clock clock.Clock,
	enqueuer JobEnqueuer,
) (*RemoveInactiveLinksJob, error) {
	return &RemoveInactiveLinksJob{
		log:      log,
		repo:     repo,
		enqueuer: enqueuer,
		clock:    clock,
	}, nil
}

func (j *RemoveInactiveLinksJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := j.log.WithContext(span.Context())

	log.Info("retrieving inactive links that can be deleted")

	// TODO This needs to be tested thoroughly
	//      1. Make sure that this doesn't return deleted links.
	//      2. Mark links as deleted when we enqueue them.
	//      3. Wrap this in a transaction?
	inactiveLinks, err := j.repo.GetInactiveLinksForExpiredAccounts(span.Context())
	if err != nil {
		log.WithError(err).Error("failed to find inactive links for expired accounts")
		return err
	}

	for i := range inactiveLinks {
		inactiveLink := inactiveLinks[i]
		linkLog := log.WithFields(logrus.Fields{
			"accountId": inactiveLink.AccountId,
			"linkId":    inactiveLink.LinkId,
		})

		linkLog.Info("enqueueing inactive link for deactivation")
		if err := j.enqueuer.EnqueueJob(span.Context(), RemoveLink, RemoveLinkArguments{
			AccountId: inactiveLink.AccountId,
			LinkId:    inactiveLink.LinkId,
		}); err != nil {
			linkLog.WithError(err).Warn("failed to enqueue link removal for inactive link on an expired account")
			continue
		}
	}

	return nil
}
