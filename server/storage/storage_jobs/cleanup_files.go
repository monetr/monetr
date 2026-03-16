package storage_jobs

import (
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
)

func CleanupFilesCron(ctx queue.Context) error {
	log := ctx.Log()

	log.DebugContext(ctx, "looking for expired files that need to be removed")

	var expiredFiles []models.File
	if err := ctx.DB().ModelContext(ctx, &expiredFiles).
		Where(`"expires_at" < ?`, ctx.Clock().Now()).
		Where(`"reconciled_at" IS NULL`).
		Select(&expiredFiles); err != nil {
		log.ErrorContext(ctx, "failed to retrieve expired filed", "err", err)
		return err
	}

	if len(expiredFiles) == 0 {
		log.DebugContext(ctx, "no expired files to remove at this time")
		return nil
	}

	log.InfoContext(ctx, "queueing expired files to be removed", "expiredFilesCount", len(expiredFiles))

	for i := range expiredFiles {
		expiredFile := expiredFiles[i]
		fileLog := log.With(
			"accountId", expiredFile.AccountId,
			"fileId", expiredFile.FileId,
		)
		fileLog.DebugContext(ctx, "queueing file to be removed")
		if err := queue.Enqueue(
			ctx,
			ctx.Enqueuer(),
			RemoveFile,
			RemoveFileArguments{
				AccountId: expiredFile.AccountId,
				FileId:    expiredFile.FileId,
			},
		); err != nil {
			fileLog.WarnContext(ctx, "failed to queue file to be removed", "err", err)
			continue
		}
	}

	return nil
}
