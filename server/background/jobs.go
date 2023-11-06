package background

import (
	"context"
	"fmt"

	"github.com/go-pg/pg/v10"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	DefaultSchedule = "0 0 0 * * *"
)

var (
	_ JobController = &BackgroundJobs{}
)

//go:generate mockgen -source=jobs.go -package=mockgen -destination=../internal/mockgen/jobs.go JobController
type (
	// JobController is an interface that can be safely provided to packages outside this one that will allow jobs to be
	// triggered manually by other events. For a job to be triggered it must have its own trigger function that accepts
	// this interface as an argument. This is to keep interaction with the background job processing to a minimum by
	// code outside this package.
	JobController interface {
		// TriggerJob is used internally to allow other areas of monetr to trigger jobs safely. This must be called by a
		// wrapping function for the specific job.
		TriggerJob(ctx context.Context, queue string, data interface{}) error
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
	configuration config.Configuration,
	db *pg.DB,
	redisPool *redis.Pool,
	publisher pubsub.Publisher,
	plaidPlatypus platypus.Platypus,
	plaidSecrets secrets.PlaidSecretsProvider,
) (*BackgroundJobs, error) {
	var enqueuer JobEnqueuer
	var processor JobProcessor

	var err error
	switch configuration.BackgroundJobs.Engine {
	case config.BackgroundJobEngineInMemory:
		panic("in-memory job engine not implemented")
	case config.BackgroundJobEngineGoCraftWork:
		enqueuer = NewGoCraftWorkJobEnqueuer(log, redisPool)
		craftProcessor := NewGoCraftWorkJobProcessor(log, configuration.BackgroundJobs, redisPool, enqueuer)
		processor = craftProcessor
	case config.BackgroundJobEngineRabbitMQ:
		panic("RabbitMQ job engine not implemented")
	case config.BackgroundJobEnginePostgreSQL:
		pqConnectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			configuration.PostgreSQL.Username,
			configuration.PostgreSQL.Password,
			configuration.PostgreSQL.Address,
			configuration.PostgreSQL.Port,
			configuration.PostgreSQL.Database,
			// TODO Need to actually implement this. TSL is used in production.
			"disable",
		)
		enqueuer, err = NewNeoqJobEnqueuerPostgres(log, pqConnectionString)
		if err != nil {
			return nil, err
		}
		neoqProcessor, err := NewNeoqJobProcessorPostgres(log, configuration.BackgroundJobs, pqConnectionString, enqueuer)
		if err != nil {
			return nil, err
		}
		processor = neoqProcessor
	default:
		return nil, errors.New("invalid background job engine specified")
	}

	jobs := []JobHandler{
		NewDeactivateLinksHandler(log, db, configuration, plaidSecrets, plaidPlatypus),
		NewProcessFundingScheduleHandler(log, db),
		NewProcessSpendingHandler(log, db),
		NewPullTransactionsHandler(log, db, plaidSecrets, plaidPlatypus, publisher),
		NewRemoveLinkHandler(log, db, publisher),
		NewRemoveTransactionsHandler(log, db),
		NewSyncPlaidHandler(log, db, plaidSecrets, plaidPlatypus, publisher),
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

func (b *BackgroundJobs) TriggerJob(ctx context.Context, queue string, data interface{}) error {
	return b.enqueuer.EnqueueJob(ctx, queue, data)
}
