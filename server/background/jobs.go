package background

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/storage"
	"github.com/sirupsen/logrus"
)

const (
	DefaultSchedule = "0 0 0 * * *"
)

var (
	_ JobController = &BackgroundJobs{}
	_ JobEnqueuer   = &BackgroundJobs{}
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=jobs.go -package=mockgen -destination=../internal/mockgen/jobs.go JobController
type (
	// JobController is an interface that can be safely provided to packages outside this one that will allow jobs to be
	// triggered manually by other events. For a job to be triggered it must have its own trigger function that accepts
	// this interface as an argument. This is to keep interaction with the background job processing to a minimum by
	// code outside this package.
	JobController interface {
		// TriggerJob is used internally to allow other areas of monetr to trigger
		// jobs safely. This must be called by a wrapping function for the specific
		// job.
		// Deprecated: Use EnqueueJobTxn instead.
		EnqueueJob(ctx context.Context, queue string, data any) error
		EnqueueJobTxn(ctx context.Context, txn pg.DBI, queue string, data any) error
	}

	BackgroundJobs struct {
		configuration config.Configuration
		jobs          []JobHandler
		enqueuer      JobEnqueuer
		processor     JobProcessor
	}
)

func NewBackgroundJobs(
	ctx context.Context,
	log *logrus.Entry,
	clock clock.Clock,
	configuration config.Configuration,
	db *pg.DB,
	publisher pubsub.Publisher,
	plaidPlatypus platypus.Platypus,
	kms secrets.KeyManagement,
	fileStorage storage.Storage,
	billing billing.Billing,
	email communication.EmailCommunication,
) (*BackgroundJobs, error) {
	var enqueuer JobEnqueuer
	var processor JobProcessor
	enqueuer = NewPostgresJobEnqueuer(
		log,
		db,
		clock,
	)
	processor = NewPostgresJobProcessor(
		log,
		clock,
		db,
		enqueuer,
	)

	jobs := []JobHandler{
		NewCalculateTransactionClustersHandler(log, db, clock),
		NewCleanupFilesHandler(log, db, clock, fileStorage, enqueuer),
		NewCleanupJobsHandler(log, db),
		NewProcessFundingScheduleHandler(log, db, clock),
		NewProcessOFXUploadHandler(log, db, clock, fileStorage, publisher, enqueuer),
		NewProcessSpendingHandler(log, db, clock),
		NewRemoveFileHandler(log, db, clock, fileStorage),
		NewRemoveLinkHandler(log, db, clock, publisher),
		NewSyncPlaidAccountsHandler(log, db, clock, kms, plaidPlatypus),
		NewSyncPlaidHandler(log, db, clock, kms, plaidPlatypus, publisher, enqueuer),
	}

	// When billing is enabled, periodically perform billing upkeep tasks.
	if configuration.Stripe.IsBillingEnabled() {
		jobs = append(jobs,
			NewDeactivateLinksHandler(log, db, clock, configuration, kms, plaidPlatypus),
			NewNotificationTrialExpiryHandler(log, db, clock, configuration, email),
			NewReconcileSubscriptionHandler(log, db, clock, publisher, billing),
			NewRemoveInactiveLinksHandler(log, db, clock, configuration, enqueuer),
		)
	}

	// Setup jobs
	for _, jobHandler := range jobs {
		if err := processor.RegisterJob(ctx, jobHandler); err != nil {
			return nil, err
		}
	}

	backgroundJobs := &BackgroundJobs{
		configuration: configuration,
		jobs:          jobs,
		enqueuer:      enqueuer,
		processor:     processor,
	}

	return backgroundJobs, nil
}

func (b *BackgroundJobs) JobNames() []string {
	names := make([]string, len(b.jobs))
	for i, job := range b.jobs {
		names[i] = job.QueueName()
	}

	return names
}

func (b *BackgroundJobs) GetTriggerableJobNames() []string {
	names := make([]string, 0, len(b.jobs))
	for _, job := range b.jobs {
		if triggerable, ok := job.(TriggerableJobHandler); ok {
			names = append(names, triggerable.QueueName())
		}
	}

	return names
}

func (b *BackgroundJobs) Start() error {
	return b.processor.Start()
}

func (b *BackgroundJobs) Close() error {
	return b.processor.Close()
}

func (b *BackgroundJobs) EnqueueJob(ctx context.Context, queue string, data any) error {
	return b.enqueuer.EnqueueJob(ctx, queue, data)
}

func (b *BackgroundJobs) EnqueueJobTxn(ctx context.Context, txn pg.DBI, queue string, data any) error {
	return b.enqueuer.EnqueueJobTxn(ctx, txn, queue, data)
}
