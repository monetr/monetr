package jobs

import (
	"fmt"
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/pkg/errors"
)

func CleanupJobsCron(ctx queue.Context) error {
	log := ctx.Log()
	log.InfoContext(ctx, "getting ready to clean up the jobs table")

	now := ctx.Clock().Now()
	cutoff := now.Add(-15 * 24 * time.Hour)
	result, err := ctx.DB().ModelContext(ctx, &models.Job{}).
		Where(`"job"."created_at" < ?`, cutoff).
		Where(`NOT ("job"."status" = ? AND "job"."priority" > ?)`, models.PendingJobStatus, now.Unix()).
		Delete()
	if err = errors.Wrap(err, "failed to cleanup old jobs from the jobs table"); err != nil {
		log.ErrorContext(ctx, "failed to cleanup", "err", err)
		return err
	}

	if affected := result.RowsAffected(); affected > 0 {
		log.InfoContext(ctx, fmt.Sprintf("deleted %d old jobs from the jobs table", affected))

		if _, err := ctx.DB().ExecContext(ctx, `VACUUM jobs;`); err != nil {
			log.ErrorContext(ctx, "failed to vacuum jobs table", "err", err)
		}

		if _, err := ctx.DB().ExecContext(ctx, `VACUUM cron_jobs;`); err != nil {
			log.ErrorContext(ctx, "failed to vacuum cron jobs table", "err", err)
		}
	} else {
		log.InfoContext(ctx, "no jobs were cleaned up from the jobs table")
	}

	return nil
}
