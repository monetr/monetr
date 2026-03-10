package background

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/pkg/errors"
)

const (
	CleanupJobs = "CleanupJobs"
)

var (
	_ ScheduledJobHandler = &CleanupJobsHandler{}
	_ JobImplementation   = &CleanupJobsJob{}
)

type CleanupJobsHandler struct {
	log *slog.Logger
	db  pg.DBI
}

func TriggerCleanupJobs(ctx context.Context, backgroundJobs JobController) error {
	return backgroundJobs.EnqueueJob(ctx, CleanupJobs, nil)
}

func NewCleanupJobsHandler(log *slog.Logger, db pg.DBI) *CleanupJobsHandler {
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
	log *slog.Logger,
	data []byte,
) error {
	span := sentry.StartSpan(ctx, "db.transaction")
	defer span.Finish()

	job := NewCleanupJobsJob(log, c.db)
	return job.Run(span.Context())
}

type CleanupJobsJob struct {
	log *slog.Logger
	db  pg.DBI
}

func NewCleanupJobsJob(log *slog.Logger, db pg.DBI) JobImplementation {
	return &CleanupJobsJob{
		log: log,
		db:  db,
	}
}

func (c *CleanupJobsJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := c.log
	log.InfoContext(span.Context(), "getting ready to clean up the jobs table")

	result, err := c.db.ModelContext(span.Context(), &models.Job{}).
		Where(`"job"."created_at" < ?`, time.Now().Add(-15*24*time.Hour)).
		Delete()
	if err = errors.Wrap(err, "failed to cleanup old jobs from the jobs table"); err != nil {
		log.ErrorContext(span.Context(), "failed to cleanup", "err", err)
		return err
	}

	if affected := result.RowsAffected(); affected > 0 {
		log.InfoContext(span.Context(), fmt.Sprintf("deleted %d old jobs from the jobs table", affected))

		if _, err := c.db.ExecContext(span.Context(), `VACUUM jobs;`); err != nil {
			log.ErrorContext(span.Context(), "failed to vacuum jobs table", "err", err)
		}

		if _, err := c.db.ExecContext(span.Context(), `VACUUM cron_jobs;`); err != nil {
			log.ErrorContext(span.Context(), "failed to vacuum cron jobs table", "err", err)
		}
	} else {
		log.InfoContext(span.Context(), "no jobs were cleaned up from the jobs table")
	}

	return nil
}

func CronCleanupJobs(ctx queue.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := ctx.Log()
	log.InfoContext(span.Context(), "getting ready to clean up the jobs table")

	result, err := ctx.DB().ModelContext(span.Context(), &models.Job{}).
		Where(`"job"."created_at" < ?`, time.Now().Add(-15*24*time.Hour)).
		Delete()
	if err = errors.Wrap(err, "failed to cleanup old jobs from the jobs table"); err != nil {
		log.ErrorContext(span.Context(), "failed to cleanup", "err", err)
		return err
	}

	if affected := result.RowsAffected(); affected > 0 {
		log.InfoContext(span.Context(), fmt.Sprintf("deleted %d old jobs from the jobs table", affected))

		if _, err := ctx.DB().ExecContext(span.Context(), `VACUUM jobs;`); err != nil {
			log.ErrorContext(span.Context(), "failed to vacuum jobs table", "err", err)
		}

		if _, err := ctx.DB().ExecContext(span.Context(), `VACUUM cron_jobs;`); err != nil {
			log.ErrorContext(span.Context(), "failed to vacuum cron jobs table", "err", err)
		}
	} else {
		log.InfoContext(span.Context(), "no jobs were cleaned up from the jobs table")
	}

	return nil
}
