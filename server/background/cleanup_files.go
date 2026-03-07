package background

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
)

const (
	CleanupFiles = "CleanupFiles"
)

var (
	_ ScheduledJobHandler = &CleanupFilesHandler{}
	_ JobImplementation   = &CleanupFilesJob{}
)

type (
	CleanupFilesHandler struct {
		log         *slog.Logger
		db          pg.DBI
		clock       clock.Clock
		fileStorage storage.Storage
		enqueuer    JobEnqueuer
	}

	CleanupFilesJob struct {
		log         *slog.Logger
		db          pg.DBI
		clock       clock.Clock
		fileStorage storage.Storage
		enqueuer    JobEnqueuer
	}
)

func NewCleanupFilesHandler(
	log *slog.Logger,
	db *pg.DB,
	clock clock.Clock,
	files storage.Storage,
	enqueuer JobEnqueuer,
) *CleanupFilesHandler {
	return &CleanupFilesHandler{
		log:         log,
		db:          db,
		clock:       clock,
		fileStorage: files,
		enqueuer:    enqueuer,
	}
}

func (CleanupFilesHandler) DefaultSchedule() string {
	// Every hour on the 15th minute
	return "0 28 * * * *"
}

func (h *CleanupFilesHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	return enqueuer.EnqueueJob(ctx, h.QueueName(), nil)
}

func (h *CleanupFilesHandler) HandleConsumeJob(
	ctx context.Context,
	log *slog.Logger,
	data []byte,
) error {
	span := sentry.StartSpan(ctx, "db.transaction")
	defer span.Finish()

	job := NewCleanupFilesJob(
		log,
		h.db,
		h.clock,
		h.fileStorage,
		h.enqueuer,
	)
	return job.Run(span.Context())
}

func (CleanupFilesHandler) QueueName() string {
	return CleanupFiles
}

func NewCleanupFilesJob(
	log *slog.Logger,
	db pg.DBI,
	clock clock.Clock,
	fileStorage storage.Storage,
	enqueuer JobEnqueuer,
) *CleanupFilesJob {
	return &CleanupFilesJob{
		log:         log,
		db:          db,
		clock:       clock,
		fileStorage: fileStorage,
		enqueuer:    enqueuer,
	}
}

func (j *CleanupFilesJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()

	log := j.log

	log.DebugContext(span.Context(), "looking for expired files that need to be removed")

	var expiredFiles []models.File
	if err := j.db.ModelContext(span.Context(), &expiredFiles).
		Where(`"expires_at" < ?`, j.clock.Now()).
		Where(`"reconciled_at" IS NULL`).
		Select(&expiredFiles); err != nil {
		log.ErrorContext(span.Context(), "failed to retrieve expired filed", "err", err)
		return err
	}

	if len(expiredFiles) == 0 {
		log.DebugContext(span.Context(), "no expired files to remove at this time")
		return nil
	}

	log.InfoContext(span.Context(), "queueing expired files to be removed", "expiredFilesCount", len(expiredFiles))

	for i := range expiredFiles {
		expiredFile := expiredFiles[i]
		fileLog := log.With(
			"accountId", expiredFile.AccountId,
			"fileId", expiredFile.FileId,
		)
		fileLog.DebugContext(span.Context(), "queueing file to be removed")
		if err := j.enqueuer.EnqueueJob(span.Context(), RemoveFile, RemoveFileArguments{
			AccountId: expiredFile.AccountId,
			FileId:    expiredFile.FileId,
		}); err != nil {
			fileLog.WarnContext(span.Context(), "failed to queue file to be removed", "err", err)
			continue
		}
	}

	return nil
}
