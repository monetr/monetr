package background

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	CleanupJobs = "CleanupJobs"
)

var (
	_ ScheduledJobHandler = &CleanupJobsHandler{}
	_ JobImplementation   = &CleanupJobsJob{}
)

type CleanupJobsHandler struct {
	log *logrus.Entry
	db  pg.DBI
}

func TriggerCleanupJobs(ctx context.Context, backgroundJobs JobController) error {
	return backgroundJobs.EnqueueJob(ctx, CleanupJobs, nil)
}

func NewCleanupJobsHandler(log *logrus.Entry, db pg.DBI) *CleanupJobsHandler {
	return &CleanupJobsHandler{
		log: log,
		db:  db,
	}
}

func (c CleanupJobsHandler) DefaultSchedule() string {
	return "0 0 8 * * *"
}

func (c CleanupJobsHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	return enqueuer.EnqueueJob(ctx, c.QueueName(), nil)
}

func (c CleanupJobsHandler) QueueName() string {
	return CleanupJobs
}

func (c *CleanupJobsHandler) HandleConsumeJob(
	ctx context.Context,
	log *logrus.Entry,
	data []byte,
) error {
	span := sentry.StartSpan(ctx, "db.transaction")
	defer span.Finish()

	job := NewCleanupJobsJob(log.WithContext(span.Context()), c.db)
	return job.Run(span.Context())
}

type CleanupJobsJob struct {
	log *logrus.Entry
	db  pg.DBI
}

func NewCleanupJobsJob(log *logrus.Entry, db pg.DBI) JobImplementation {
	return &CleanupJobsJob{
		log: log,
		db:  db,
	}
}

func (c *CleanupJobsJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := c.log.WithContext(span.Context())
	log.Info("getting ready to clean up the jobs table")

	result, err := c.db.ModelContext(span.Context(), &models.Job{}).
		Where(`"job"."created_at" < ?`, time.Now().Add(-15*24*time.Hour)).
		Delete()
	if err = errors.Wrap(err, "failed to cleanup old jobs from the jobs table"); err != nil {
		log.WithError(err).Errorf("failed to cleanup")
		return err
	}

	if affected := result.RowsAffected(); affected > 0 {
		log.WithFields(logrus.Fields{
			"deleted": affected,
		}).Info("deleted old jobs from the jobs table")

		if _, err := c.db.ExecContext(span.Context(), `VACUUM jobs;`); err != nil {
			log.WithError(err).Error("failed to vacuum jobs table")
		}

		if _, err := c.db.ExecContext(span.Context(), `VACUUM cron_jobs;`); err != nil {
			log.WithError(err).Error("failed to vacuum cron jobs table")
		}
	} else {
		log.Info("no jobs were cleaned up from the jobs table")
	}

	return nil
}
