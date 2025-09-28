package background

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	ProcessCAMTUpload = "ProcessCAMTUpload"
)

type (
	ProcessCAMTUploadHandler struct {
		log          *logrus.Entry
		db           *pg.DB
		publisher    pubsub.Publisher
		files        storage.Storage
		enqueuer     JobEnqueuer
		unmarshaller JobUnmarshaller
		clock        clock.Clock
	}

	ProcessCAMTUploadArguments struct {
		AccountId           ID[Account]           `json:"accountId"`
		LinkId              ID[Link]              `json:"linkId"`
		TransactionImportId ID[TransactionImport] `json:"transactionImportId"`
	}

	ProcessCAMTUploadJob struct {
		args ProcessCAMTUploadArguments

		log       *logrus.Entry
		repo      repository.BaseRepository
		files     storage.Storage
		publisher pubsub.Publisher
		enqueuer  JobEnqueuer
		clock     clock.Clock
		timezone  *time.Location
	}
)

func NewProcessCAMTUploadHandler(
	log *logrus.Entry,
	db *pg.DB,
	clock clock.Clock,
	files storage.Storage,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
) *ProcessCAMTUploadHandler {
	return &ProcessCAMTUploadHandler{
		log:          log,
		db:           db,
		publisher:    publisher,
		files:        files,
		enqueuer:     enqueuer,
		unmarshaller: DefaultJobUnmarshaller,
		clock:        clock,
	}
}

func (ProcessCAMTUploadHandler) QueueName() string {
	return ProcessCAMTUpload
}

func (h *ProcessCAMTUploadHandler) HandleConsumeJob(
	ctx context.Context,
	inLog *logrus.Entry,
	data []byte,
) error {
	var args ProcessCAMTUploadArguments
	if err := errors.Wrap(h.unmarshaller(data, &args), "failed to unmarshal arguments"); err != nil {
		crumbs.Error(ctx, "Failed to unmarshal arguments for Processing CAMT.053 Upload job.", "job", map[string]any{
			"data": data,
		})
		return err
	}

	crumbs.IncludeUserInScope(ctx, args.AccountId)

	log := inLog.WithFields(logrus.Fields{
		"accountId":           args.AccountId,
		"linkId":              args.LinkId,
		"transactionImportId": args.TransactionImportId,
	})

	return h.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		log := log.WithContext(span.Context())
		repo := repository.NewRepositoryFromSession(
			h.clock,
			"user_system",
			args.AccountId,
			txn,
			log,
		)

		job, err := NewProcessCAMTUploadJob(
			log, repo, h.clock, h.files, h.publisher, h.enqueuer, args,
		)
		if err != nil {
			return err
		}

		return job.Run(span.Context())
	})
}

func NewProcessCAMTUploadJob(
	log *logrus.Entry,
	repo repository.BaseRepository,
	clock clock.Clock,
	files storage.Storage,
	publisher pubsub.Publisher,
	enqueuer JobEnqueuer,
	args ProcessCAMTUploadArguments,
) (*ProcessCAMTUploadJob, error) {
	return &ProcessCAMTUploadJob{
		args:      args,
		log:       log,
		repo:      repo,
		files:     files,
		publisher: publisher,
		enqueuer:  enqueuer,
		clock:     clock,
	}, nil
}

func (j *ProcessCAMTUploadJob) Run(ctx context.Context) error {
	span := sentry.StartSpan(ctx, "job.exec")
	defer span.Finish()
	crumbs.IncludeUserInScope(span.Context(), j.args.AccountId)

	log := j.log.WithContext(span.Context())

	account, err := j.repo.GetAccount(span.Context())
	if err != nil {
		log.WithError(err).Error("failed to retrieve account for job")
		return err
	}

	j.timezone, err = account.GetTimezone()
	if err != nil {
		log.WithError(err).Warn("failed to get account's time zone, defaulting to UTC")
		j.timezone = time.UTC
	}

	return nil
}

func (j *ProcessCAMTUploadJob) loadFile(ctx context.Context) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	return nil
}
