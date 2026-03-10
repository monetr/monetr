package background

import (
	"context"
	"log/slog"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
)

const (
	RemoveFileName = "RemoveFile"
)

var (
	_ JobHandler = &RemoveFileHandler{}
)

type (
	RemoveFileHandler struct {
		log          *slog.Logger
		db           *pg.DB
		files        storage.Storage
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	RemoveFileArguments struct {
		AccountId ID[Account] `json:"accountId"`
		FileId    ID[File]    `json:"fileId"`
	}

	RemoveFileJob struct {
		args  RemoveFileArguments
		log   *slog.Logger
		repo  repository.BaseRepository
		files storage.Storage
		clock clock.Clock
	}
)

func NewRemoveFileHandler(
	log *slog.Logger,
	db *pg.DB,
	clock clock.Clock,
	files storage.Storage,
) *RemoveFileHandler {
	return &RemoveFileHandler{
		log:          log,
		db:           db,
		files:        files,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

func (h *RemoveFileHandler) QueueName() string {
	return RemoveFileName
}

func (h *RemoveFileHandler) HandleConsumeJob(
	ctx context.Context,
	log *slog.Logger,
	data []byte,
) error {
	var args RemoveFileArguments
	if err := errors.Wrap(h.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Remove File job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return h.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.With(
			"accountId", args.AccountId,
			"fileId", args.FileId,
		)
		repo := repository.NewRepositoryFromSession(
			h.clock,
			"user_system",
			args.AccountId,
			txn,
			log,
		)

		job, err := NewRemoveFileJob(
			log,
			repo,
			h.clock,
			h.files,
			args,
		)
		if err != nil {
			return err
		}

		return job.Run(span.Context())
	})
}

func NewRemoveFileJob(
	log *slog.Logger,
	repo repository.BaseRepository,
	clock clock.Clock,
	fileStorage storage.Storage,
	args RemoveFileArguments,
) (*RemoveFileJob, error) {
	return &RemoveFileJob{
		args:  args,
		log:   log,
		repo:  repo,
		files: fileStorage,
		clock: clock,
	}, nil
}

func (j *RemoveFileJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := j.log.With(
		"accountId", j.args.AccountId,
		"fileId", j.args.FileId,
	)

	file, err := j.repo.GetFile(span.Context(), j.args.FileId)
	if err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve file from database", "err", err)
		return err
	}

	if file.ReconciledAt != nil {
		log.InfoContext(span.Context(), "file is already deleted")
		return nil
	}

	log.DebugContext(span.Context(), "removing file")
	if err = j.files.Remove(span.Context(), *file); err != nil {
		log.ErrorContext(span.Context(), "failed to remove file", "err", err)
	}

	now := j.clock.Now()
	file.ReconciledAt = &now

	if file.DeletedAt == nil {
		file.DeletedAt = &now
	}

	if err := j.repo.UpdateFile(span.Context(), file); err != nil {
		log.ErrorContext(span.Context(), "failed to update file's reconciled at", "err", err)
		return err
	}

	log.DebugContext(span.Context(), "file successfully removed")
	return nil
}

func RemoveFile(ctx queue.Context, args RemoveFileArguments) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := ctx.Log().With(
		"accountId", args.AccountId,
		"fileId", args.FileId,
	)

	return ctx.DB().RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		repo := repository.NewRepositoryFromSession(
			clock.New(), // TODO!
			"user_system",
			args.AccountId,
			txn,
			log,
		)

		file, err := repo.GetFile(span.Context(), args.FileId)
		if err != nil {
			log.ErrorContext(span.Context(), "failed to retrieve file from database", "err", err)
			return err
		}

		if file.ReconciledAt != nil {
			log.InfoContext(span.Context(), "file is already deleted")
			return nil
		}

		log.DebugContext(span.Context(), "removing file")
		if err = ctx.Storage().Remove(span.Context(), *file); err != nil {
			log.ErrorContext(span.Context(), "failed to remove file", "err", err)
		}

		now := time.Now()
		file.ReconciledAt = &now

		if file.DeletedAt == nil {
			file.DeletedAt = &now
		}

		if err := repo.UpdateFile(span.Context(), file); err != nil {
			log.ErrorContext(span.Context(), "failed to update file's reconciled at", "err", err)
			return err
		}

		log.DebugContext(span.Context(), "file successfully removed")
		return nil
	})

}
