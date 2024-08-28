package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
	"github.com/sirupsen/logrus"
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
		log         *logrus.Entry
		db          pg.DBI
		clock       clock.Clock
		fileStorage storage.Storage
		enqueuer    JobEnqueuer
	}

	CleanupFilesJob struct {
		log         *logrus.Entry
		db          pg.DBI
		clock       clock.Clock
		fileStorage storage.Storage
		enqueuer    JobEnqueuer
	}
)

func NewCleanupFilesHandler(
	log *logrus.Entry,
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
	log *logrus.Entry,
	data []byte,
) error {
	span := sentry.StartSpan(ctx, "db.transaction")
	defer span.Finish()

	job := NewCleanupFilesJob(
		log.WithContext(span.Context()),
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
	log *logrus.Entry,
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

	log := j.log.WithContext(span.Context())

	log.Debug("looking for expired files that need to be removed")

	var expiredFiles []models.File
	if err := j.db.ModelContext(span.Context(), &expiredFiles).
		Where(`"expires_at" < ?`, j.clock.Now()).
		Where(`"reconciled_at" IS NULL`).
		Select(&expiredFiles); err != nil {
		log.WithError(err).Error("failed to retrieve expired filed")
		return err
	}

	if len(expiredFiles) == 0 {
		log.Debug("no expired files to remove at this time")
		return nil
	}

	log.WithField("expiredFilesCount", len(expiredFiles)).
		Info("queueing expired files to be removed")

	for i := range expiredFiles {
		expiredFile := expiredFiles[i]
		fileLog := log.WithFields(logrus.Fields{
			"accountId": expiredFile.AccountId,
			"fileId":    expiredFile.FileId,
		})
		fileLog.Debug("queueing file to be removed")
		if err := j.enqueuer.EnqueueJob(span.Context(), RemoveFile, RemoveFileArguments{
			AccountId: expiredFile.AccountId,
			FileId:    expiredFile.FileId,
		}); err != nil {
			fileLog.WithError(err).Warn("failed to queue file to be removed")
			continue
		}
	}

	return nil
}
