package storage_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
)

func CleanupFilesCron(ctx queue.Context) error {
	log := ctx.Log()

	log.DebugContext(ctx, "looking for expired files that need to be removed")

	var expiredFiles []models.File
	if err := ctx.DB().NewSelect().Model(&expiredFiles).
		Where(`"expires_at" < ?`, ctx.Clock().Now()).
		Where(`"reconciled_at" IS NULL`).
		Scan(ctx); err != nil {
		log.ErrorContext(ctx, "failed to retrieve expired filed", "err", err)
		return err
	}

	if len(expiredFiles) == 0 {
		log.DebugContext(ctx, "no expired files to remove at this time")
		return nil
	}

	log.InfoContext(ctx, "queueing expired files to be removed", "expiredFilesCount", len(expiredFiles))

	if err := queue.BulkEnqueue(
		ctx,
		ctx.Enqueuer(),
		RemoveFile,
		myownsanity.Map(
			expiredFiles,
			func(expiredFile models.File) RemoveFileArguments {
				return RemoveFileArguments{
					AccountId: expiredFile.AccountId,
					FileId:    expiredFile.FileId,
				}
			}),
	); err != nil {
		log.WarnContext(ctx, "failed to queue files to be removed", "err", err)
		crumbs.Warn(ctx, "Failed to queue files to be removed", "job", map[string]any{
			"error": err,
		})
		return err
	}

	return nil
}
