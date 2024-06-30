package background

import (
	"context"
	"fmt"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	RemoveFile = "RemoveFile"
)

var (
	_ JobHandler = &RemoveFileHandler{}
)

type (
	RemoveFileHandler struct {
		log          *logrus.Entry
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
		log   *logrus.Entry
		repo  repository.BaseRepository
		files storage.Storage
		clock clock.Clock
	}
)

func NewRemoveFileHandler(
	log *logrus.Entry,
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
	return RemoveFile
}

func (h *RemoveFileHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	var args RemoveFileArguments
	if err := errors.Wrap(h.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Remove File job.", "job", map[string]interface{}{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	return h.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := h.log.WithContext(span.Context())
		repo := repository.NewRepositoryFromSession(h.clock, "user_system", args.AccountId, txn)

		fmt.Sprint(log, repo)

		return nil
	})
}

func NewRemoveFileJob(
	log *logrus.Entry,
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

	log := j.log.
		WithContext(span.Context()).
		WithFields(logrus.Fields{
			"accountId": j.args.AccountId,
			"fileId":    j.args.FileId,
		})

	file, err := j.repo.GetFile(span.Context(), j.args.FileId)
	if err != nil {
		log.WithError(err).Error("failed to retrieve file from database")
		return err
	}

	if file.DeletedAt != nil {
		log.Info("file is already deleted")
		return nil
	}

	log = log.WithField("uri", file.BlobUri)
	log.Debug("removing file")
	if err = j.files.Remove(span.Context(), file.BlobUri); err != nil {
		log.WithError(err).Error("failed to remove file")
	}

	// TODO Reconciled at and deleted at

	return nil
}
