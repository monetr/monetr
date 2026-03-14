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

	result, err := ctx.DB().ModelContext(ctx, &models.Job{}).
		Where(`"job"."created_at" < ?`, ctx.Clock().Now().Add(-15*24*time.Hour)).
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
