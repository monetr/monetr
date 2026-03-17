package jobs

import (
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/pkg/errors"
)

const (
	// stalledJobThreshold is the amount of time a job can be in a processing
	// state before it is considered stalled. If a job has been processing for
	// longer than this duration, it will be recovered by either retrying it or
	// marking it as failed.
	stalledJobThreshold = 10 * time.Minute

	// maxAttempts must match the constant in the queue package.
	maxAttempts = 5
)

func RecoverStalledJobsCron(ctx queue.Context) error {
	log := ctx.Log()
	log.InfoContext(ctx, "checking for stalled jobs")

	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		now := ctx.Clock().Now()
		cutoff := now.Add(-stalledJobThreshold)

		// Any jobs that have a started at more than 10 minutes ago and are in a
		// processing status have likely timed out. We want to requeue the ones that
		// have remaining attempts left.
		retried, err := ctx.DB().ModelContext(ctx, new(models.Job)).
			Set(`"status" = ?`, models.PendingJobStatus).
			Set(`"attempt" = "attempt" + 1`).
			Set(`"started_at" = NULL`).
			Set(`"updated_at" = ?`, now).
			Where(
				`"job_id" IN (?)`,
				ctx.DB().Model(new(models.Job)).
					Column("job_id").
					Where(`"status" = ?`, models.ProcessingJobStatus).
					Where(`"started_at" < ?`, cutoff).
					Where(`"attempt" < ?`, maxAttempts).
					For(`UPDATE SKIP LOCKED`),
			).
			Update()
		if err != nil {
			return errors.Wrap(err, "failed to recover stalled jobs")
		}

		// Mark stalled jobs that have exhausted all attempts as failed.
		failed, err := ctx.DB().ModelContext(ctx, new(models.Job)).
			Set(`"status" = ?`, models.FailedJobStatus).
			Set(`"completed_at" = ?`, now).
			Set(`"updated_at" = ?`, now).
			Where(
				`"job_id" IN (?)`,
				ctx.DB().Model(new(models.Job)).
					Column("job_id").
					Where(`"status" = ?`, models.ProcessingJobStatus).
					Where(`"started_at" < ?`, cutoff).
					Where(`"attempt" >= ?`, maxAttempts).
					For(`UPDATE SKIP LOCKED`),
			).
			Update()
		if err != nil {
			return errors.Wrap(err, "failed to mark exhausted stalled jobs as failed")
		}

		retriedCount := retried.RowsAffected()
		failedCount := failed.RowsAffected()
		if retriedCount > 0 || failedCount > 0 {
			log.WarnContext(ctx, "recovered stalled jobs",
				"retried", retriedCount,
				"failed", failedCount,
			)
		} else {
			log.InfoContext(ctx, "no stalled jobs found")
		}

		return nil
	})
}
