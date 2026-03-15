package storage_jobs

import (
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
)

type RemoveFileArguments struct {
	AccountId models.ID[models.Account] `json:"accountId"`
	FileId    models.ID[models.File]    `json:"fileId"`
}

func RemoveFile(ctx queue.Context, args RemoveFileArguments) error {
	return ctx.RunInTransaction(ctx, func(ctx queue.Context) error {
		crumbs.IncludeUserInScope(ctx, args.AccountId)
		log := ctx.Log().With(
			"accountId", args.AccountId,
			"fileId", args.FileId,
		)

		repo := repository.NewRepositoryFromSession(
			ctx.Clock(),
			"user_system",
			args.AccountId,
			ctx.DB(),
			log,
		)

		file, err := repo.GetFile(ctx, args.FileId)
		if err != nil {
			log.ErrorContext(ctx, "failed to retrieve file from database", "err", err)
			return err
		}

		if file.ReconciledAt != nil {
			log.InfoContext(ctx, "file is already deleted")
			return nil
		}

		now := ctx.Clock().Now()
		file.ReconciledAt = &now

		if file.DeletedAt == nil {
			file.DeletedAt = &now
		}

		// Mark the file as removed before trying to do anything in the storage
		// layer. This way if it fails we don't remove the file from the storage
		// layer before doing this.
		if err := repo.UpdateFile(ctx, file); err != nil {
			log.ErrorContext(ctx, "failed to update file's reconciled at", "err", err)
			return err
		}

		log.DebugContext(ctx, "removing file")
		if err = ctx.Storage().Remove(ctx, *file); err != nil {
			log.ErrorContext(ctx, "failed to remove file", "err", err)
			return err
		}

		log.DebugContext(ctx, "file successfully removed")
		return nil
	})
}
