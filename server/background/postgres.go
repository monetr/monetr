package background

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"sort"
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

const (
	numberOfPostgresQueueWorkers = 4
)

const (
	postgresJobQueueUninitialized = 0
	postgresJobQueueRunning       = 1
	postgresJobQueueStopped       = 2
)

type postgresJobEnqueuer struct {
	log     *logrus.Entry
	db      pg.DBI
	clock   clock.Clock
	marshal JobMarshaller
}

func NewPostgresJobEnqueuer(
	log *logrus.Entry,
	db pg.DBI,
	clock clock.Clock,
) *postgresJobEnqueuer {
	return &postgresJobEnqueuer{
		log:     log,
		db:      db,
		clock:   clock,
		marshal: DefaultJobMarshaller,
	}
}

func (p *postgresJobEnqueuer) EnqueueJob(ctx context.Context, queue string, arguments interface{}) error {
	span := sentry.StartSpan(ctx, "topic.send")
	defer span.Finish()
	span.Description = "postgres Enqueue"
	span.SetTag("queue", queue)
	span.Data = map[string]interface{}{
		"queue":     queue,
		"arguments": arguments,
	}

	crumbs.Debug(
		span.Context(),
		"Enqueueing job for background processing",
		map[string]interface{}{
			"queue":     queue,
			"arguments": arguments,
		},
	)

	log := p.log.WithContext(span.Context()).
		WithField("queue", queue)

	log.Debug("enqueuing job to be run")

	encodedArguments, err := p.marshal(arguments)
	if err = errors.Wrap(err, "failed to marshal arguments"); err != nil {
		return err
	}

	timestamp := p.clock.Now().UTC()

	var signature string
	{ // Build the signature using a hash of the arguments and a truncated timestamp.
		truncatedTimestamp := timestamp.Truncate(time.Second)
		signatureBuilder := fnv.New32()
		signatureBuilder.Write(encodedArguments)
		signatureBuilder.Write([]byte(truncatedTimestamp.String()))
		signature = hex.EncodeToString(signatureBuilder.Sum(nil))
	}

	job := models.Job{
		Queue:       queue,
		Signature:   signature,
		Input:       encodedArguments,
		Output:      nil,
		Status:      models.PendingJobStatus,
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
		StartedAt:   nil,
		CompletedAt: nil,
	}
	_, err = p.db.ModelContext(span.Context(), &job).Insert(&job)
	if err != nil {
		// TODO Check to see if this is a conflict error, we don't want to enqueue
		// the same job if it is already in the queue. It will conflict on the
		// signature column if that is the case.
		return errors.Wrap(err, "failed to enqueue job for postgres")
	}

	log.WithFields(logrus.Fields{
		"jobId":     job.JobId,
		"signature": signature,
	}).Debug("successfully enqueued job")

	return nil
}

type postgresJobFunction func(ctx context.Context, job *models.Job) error

type postgresJobProcessor struct {
	state            uint32
	availableThreads chan struct{}
	shutdownThreads  []chan chan struct{}
	dispatch         chan *models.Job
	trigger          chan struct{}
	cronJobQueues    []ScheduledJobHandler
	queues           []string
	registeredJobs   map[string]postgresJobFunction
	jobQuery         *pg.Query
	clock            clock.Clock
	configuration    config.BackgroundJobs
	log              *logrus.Entry
	db               pg.DBI
	enqueuer         JobEnqueuer
	marshal          JobMarshaller
}

func NewPostgresJobProcessor(
	log *logrus.Entry,
	configuration config.BackgroundJobs,
	clock clock.Clock,
	db pg.DBI,
	enqueuer JobEnqueuer,
) *postgresJobProcessor {
	return &postgresJobProcessor{
		shutdownThreads: []chan chan struct{}{},
		dispatch:        make(chan *models.Job),
		trigger:         make(chan struct{}),
		cronJobQueues:   []ScheduledJobHandler{},
		queues:          []string{},
		registeredJobs:  map[string]postgresJobFunction{},
		clock:           clock,
		configuration:   configuration,
		log:             log,
		db:              db,
		enqueuer:        enqueuer,
		marshal:         DefaultJobMarshaller,
	}
}

func (p *postgresJobProcessor) RegisterJob(
	ctx context.Context,
	handler JobHandler,
) error {
	if atomic.LoadUint32(&p.state) != postgresJobQueueUninitialized {
		return errors.New("jobs cannot be added to the postgres job processor after it has been started or closed")
	}

	log := p.log.WithContext(ctx).WithField("job", handler.QueueName())
	log.Trace("registering job handler")

	if _, ok := p.registeredJobs[handler.QueueName()]; ok {
		return errors.Errorf(
			"job has already been registered: %s",
			handler.QueueName(),
		)
	}

	p.registeredJobs[handler.QueueName()] = p.buildJobExecutor(handler)
	p.queues = append(p.queues, handler.QueueName())

	if p.configuration.Scheduler == config.BackgroundJobSchedulerInternal {
		if scheduledJob, ok := handler.(ScheduledJobHandler); ok {
			schedule := scheduledJob.DefaultSchedule()
			log.WithField("schedule", schedule).
				Trace("job will be run on a schedule automatically")
			p.cronJobQueues = append(p.cronJobQueues, scheduledJob)
		}
	}

	return nil
}

func (p *postgresJobProcessor) Start() error {
	if !atomic.CompareAndSwapUint32(
		&p.state,
		postgresJobQueueUninitialized,
		postgresJobQueueRunning,
	) {
		return errors.New("postgres job processor is either already started, or is in an invalid state")
	}

	if len(p.registeredJobs) == 0 {
		// Reset the state so start could be called again.
		atomic.StoreUint32(&p.state, postgresJobQueueUninitialized)
		return errors.New("cannot start processor with no jobs registered")
	}

	if err := p.prepareCronJobTable(); err != nil {
		return err
	}

	// Setup the job query to be used by the consumer. It is built to only
	// retrieve jobs for the queues that have been reigstered.
	p.jobQuery = p.db.Model(new(models.Job)).
		Column("job_id").
		Where(`"status" = ?`, models.PendingJobStatus).
		WhereIn(`"queue" IN (?)`, p.queues).
		Order(`created_at ASC`).
		For(`UPDATE SKIP LOCKED`).
		Limit(1)

	numberOfWorkers := numberOfPostgresQueueWorkers
	// Number of threads is the number of workers plus the number of other things we kick off to consume things. At least
	// one for regular jobs, and then another for cron jobs if there are any crons enabled.
	numberOfThreads := numberOfWorkers + 1

	// If we are using cron jobs then we will kick off another thread.
	if len(p.cronJobQueues) > 0 {
		numberOfThreads += 1
	}

	p.availableThreads = make(chan struct{}, numberOfWorkers)
	p.shutdownThreads = make([]chan chan struct{}, numberOfThreads)
	// Start the worker threads.
	for i := 0; i < numberOfWorkers; i++ {
		p.shutdownThreads[i] = make(chan chan struct{})
		go p.worker(p.shutdownThreads[i])
	}

	// If there are any cron jobs registered then start the cron consumer.
	if len(p.cronJobQueues) > 0 {
		// Use number of
		p.shutdownThreads[numberOfThreads-2] = make(chan chan struct{})
		go p.cronConsumer(p.shutdownThreads[numberOfThreads-2])
	}

	// Start the consumer thread.
	p.shutdownThreads[numberOfThreads-1] = make(chan chan struct{})
	go p.backgroundConsumer(p.shutdownThreads[numberOfThreads-1])

	return nil
}

// getCronJobQueueName is used to make the cron job queue names consistent
// throughout this code.
func (p *postgresJobProcessor) getCronJobQueueName(
	handler ScheduledJobHandler,
) string {
	return fmt.Sprintf("%s::CronJob", handler.QueueName())
}

func (p *postgresJobProcessor) prepareCronJobTable() error {
	cronJobQueueNames := make([]string, 0, len(p.cronJobQueues))
	for _, cronJob := range p.cronJobQueues {
		cronJobQueueNames = append(cronJobQueueNames, p.getCronJobQueueName(cronJob))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		{ // Clean up cron jobs that are not registered.
			result, err := tx.ModelContext(ctx, new(models.CronJob)).
				WhereIn(`"queue" NOT IN (?)`, cronJobQueueNames).
				Delete()
			if err != nil {
				return errors.Wrap(err, "failed to clean up old cron jobs from postgres")
			} else if affected := result.RowsAffected(); affected > 0 {
				p.log.
					Infof("removed %d outdated cron job(s) from postgres", affected)
			} else {
				p.log.
					Debug("no outdated cron jobs were removed")
			}
		}

		for _, cronJob := range p.cronJobQueues {
			queue := p.getCronJobQueueName(cronJob)
			schedule := cronJob.DefaultSchedule()

			cronSchedule, err := cron.Parse(schedule)
			if err != nil {
				return errors.Wrapf(err, "failed to parse cron schedule for job: %s - %s", queue, schedule)
			}
			nextRunAt := cronSchedule.Next(p.clock.Now().UTC())

			log := p.log.WithFields(logrus.Fields{
				"queue":    queue,
				"schedule": schedule,
				"next":     nextRunAt,
			})
			log.Debug("upserting cron job into postgres")
			result, err := tx.ModelContext(ctx, &models.CronJob{
				Queue:        queue,
				CronSchedule: schedule,
				LastRunAt:    nil,
				NextRunAt:    nextRunAt,
			}).
				OnConflict(`("queue") DO UPDATE`).
				Set(`"cron_schedule" = ?`, cronJob.DefaultSchedule()).
				Insert()
			if err != nil {
				return errors.Wrapf(err, "failed to provision cron job: %s", cronJob)
			}

			if result.RowsAffected() == 0 {
				log.Trace("cron job already exists and did not need updating")
			} else {
				log.Debug("cron job was updated")
			}
		}

		return nil
	})
}

func (p *postgresJobProcessor) Close() error {
	if !atomic.CompareAndSwapUint32(
		&p.state,
		postgresJobQueueRunning,
		postgresJobQueueStopped,
	) {
		return errors.New("postgres job processor is either already closed, or is in an invalid state")
	}

	p.log.Info("shutting down postgres job processor")

	// Create a channel buffer with the number of messages we need to send to all the worker threads.
	shutdownChannel := make(chan struct{}, len(p.shutdownThreads))
	// Then send the shutdown channel to each worker thread as a "promise".
	for i := range p.shutdownThreads {
		p.shutdownThreads[i] <- shutdownChannel
	}

	// Then wait for the expected number of responses.
	// TODO Add timeout
	for i := 0; i < len(p.shutdownThreads); i++ {
		<-shutdownChannel
	}
	close(shutdownChannel)
	for _, channel := range p.shutdownThreads {
		close(channel)
	}

	p.log.Info("postgres job processor shut down complete")

	return nil
}

func (p *postgresJobProcessor) backgroundConsumer(shutdown chan chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		// Block if there are no available threads but allow the consumer to be
		// shutdown while we are waiting.
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down job consumer")
			promise <- struct{}{}
			ticker.Stop()
			return
		case <-p.availableThreads:
			// This receive blocks if there are no available threads.
		}

		// Before we even tick, try to consume a job.
		job, err := p.consumeJobMaybe()
		if err != nil {
			p.log.WithError(err).Error("failed to consume job")
		} else if job != nil {
			p.log.WithFields(logrus.Fields{
				"jobId": job.JobId,
				"queue": job.Queue,
			}).Debug("successfully consumed job, dispatching to worker thread")
			p.dispatch <- job
		} else {
			// If we did not retrieve a job at all then we need to put our "hold" on an available thread back in the channel
			// this way a thread can still be consumed on the next loop.
			p.availableThreads <- struct{}{}
		}

		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down job consumer")
			promise <- struct{}{}
			ticker.Stop()
			return
		case <-p.trigger:
			continue
		case <-ticker.C:
			continue
		}
	}
}

func (p *postgresJobProcessor) consumeJobMaybe() (*models.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	p.log.Trace("attempting to consume job from the queue")
	var job models.Job
	result, err := p.db.ModelContext(ctx, &job).
		Set(`"status" = ?`, models.ProcessingJobStatus).
		Set(`"started_at" = ?`, p.clock.Now()).
		Where(`"job_id" = (?)`, p.jobQuery).
		Returning("*").
		Update(&job)
	if err != nil {
		if err == pg.ErrNoRows {
			p.log.Trace("no job consumed")
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to consume job")
	}

	if result.RowsAffected() == 0 {
		p.log.Trace("no job consumed")
		return nil, nil
	}

	p.log.WithFields(logrus.Fields{
		"jobId": job.JobId,
		"queue": job.Queue,
	}).Trace("found job")

	return &job, nil
}

func (p *postgresJobProcessor) cronConsumer(shutdown chan chan struct{}) {
	type cronJobTracker struct {
		handler   ScheduledJobHandler
		schedule  cron.Schedule
		queueName string
		next      time.Time
	}

	crons := make([]cronJobTracker, len(p.cronJobQueues))
	for i, queue := range p.cronJobQueues {
		schedule, err := cron.Parse(queue.DefaultSchedule())
		if err != nil {
			panic("cron schedule is not valid, job processor should have have started")
		}

		crons[i] = cronJobTracker{
			handler:   p.cronJobQueues[i],
			schedule:  schedule,
			queueName: p.getCronJobQueueName(queue),
			next:      schedule.Next(p.clock.Now()),
		}
	}

	for {
		// Sort the crons by their next time a cron job happens.
		sort.Slice(crons, func(i, j int) bool {
			return crons[i].next.Before(crons[j].next)
		})

		// Grab the next cron that will happen, we are going to watch for this one.
		nextJob := crons[0]

		// TODO What happens if sleep is negative or 0
		// How long do we need to wait for this cron job?
		sleep := nextJob.next.Sub(p.clock.Now())

		// Create a timer.
		timer := time.NewTimer(sleep)
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down cron job consumer")
			promise <- struct{}{}
			timer.Stop()
			return
		case <-timer.C:
			// Bump the cron we just did.
			nextTimestamp := nextJob.schedule.Next(p.clock.Now())
			crons[0].next = nextTimestamp
			log := p.log.WithFields(logrus.Fields{
				"queue": nextJob.queueName,
			})

			consumedCronJob, err := p.consumeCronMaybe(nextJob.queueName, nextTimestamp)
			if err != nil {
				log.WithError(err).Error("failed to consume cron job")
				continue
			}

			if consumedCronJob == nil {
				log.Trace("did not consume cron job")
				continue
			}

			log.Trace("consumed cron job")
			if err := nextJob.handler.EnqueueTriggeredJob(context.Background(), p.enqueuer); err != nil {
				log.WithError(err).Error("failed to enqueue cron job to be executed")
				continue
			}

			// If possible, try to trigger the consumption of the cron job locally.
			select {
			case p.trigger <- struct{}{}:
				log.Trace("successfully triggered immediate job consumption")
			default:
				log.Trace("trigger queue was full, cron job processing may be slightly delayed")
			}
		}
	}
}

func (p *postgresJobProcessor) consumeCronMaybe(queue string, next time.Time) (*models.CronJob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	subQuery := p.db.ModelContext(ctx, new(models.CronJob)).
		Column("queue").
		Where(`"queue" = ?`, queue).
		Where(`"next_run_at" <= ?`, next).
		For(`UPDATE SKIP LOCKED`).
		Limit(1)

	var cronJob models.CronJob
	result, err := p.db.ModelContext(ctx, &cronJob).
		Set(`"last_run_at" = "next_run_at"`).
		Set(`"next_run_at" = ?`, next).
		Where(`"queue" = (?)`, subQuery).
		Update(&cronJob)
	if err != nil {
		return nil, errors.Wrap(err, "failed to consume job")
	}

	if result.RowsAffected() == 0 {
		return nil, nil
	}

	return &cronJob, nil
}

func (p *postgresJobProcessor) worker(shutdown chan chan struct{}) {
	for {
		// Tell the consumer that a worker is available.
		p.availableThreads <- struct{}{}
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down worker thread")
			promise <- struct{}{}
			return
		case job := <-p.dispatch:
			log := p.log.WithFields(logrus.Fields{
				"jobId": job.JobId,
				"queue": job.Queue,
			})
			log.Info("processing job")

			executor, ok := p.registeredJobs[job.Queue]
			if !ok {
				panic(fmt.Sprintf(
					"could not execute job with queue name: %s no handler registered",
					job.Queue,
				))
			}

			// Execute the job with a 1 minute timeout.
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			if err := executor(ctx, job); err != nil {
				log.WithError(err).Error("failed to execute job")
			}
			cancel()
		}
	}
}

func (p *postgresJobProcessor) markJobStatus(
	ctx context.Context,
	job *models.Job,
	jobError error,
) {
	log := p.log.
		WithContext(ctx).
		WithFields(logrus.Fields{
			"jobId": job.JobId,
			"queue": job.Queue,
		})

	if jobError != nil {
		log.Debug("marking job as failed")
		_, err := p.db.ModelContext(ctx, job).
			Set(`"completed_at" = ?`, p.clock.Now().UTC()).
			Set(`"status" = ?`, models.FailedJobStatus).
			WherePK().
			Update(&job)
		if err != nil {
			log.WithError(err).Error("failed to update job status")
		}

		return
	}

	log.Debug("marking job as complete")
	_, err := p.db.ModelContext(ctx, job).
		Set(`"completed_at" = ?`, p.clock.Now().UTC()).
		Set(`"status" = ?`, models.CompletedJobStatus).
		WherePK().
		Update(&job)
	if err != nil {
		log.WithError(err).Error("failed to update job status")
	}
}

func (p *postgresJobProcessor) buildJobExecutor(
	handler JobHandler,
) postgresJobFunction {
	return func(ctx context.Context, job *models.Job) (err error) {
		// We want to have sentry tracking jobs as they are being processed. In order to do this we need to inject a
		// sentry hub into the context and create a new span using that context.
		highContext := sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())
		span := sentry.StartSpan(
			highContext,
			"topic.process",
			sentry.TransactionName(handler.QueueName()),
		)
		span.Description = handler.QueueName()
		jobLog := p.log.WithContext(span.Context())
		hub := sentry.GetHubFromContext(span.Context())
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("queue", handler.QueueName())
		})

		defer func() {
			if panicErr := recover(); panicErr != nil {
				jobLog.Error("panic while processing job")
				if hub != nil {
					hub.RecoverWithContext(span.Context(), panicErr)
					hub.ConfigureScope(func(scope *sentry.Scope) {
						scope.SetLevel(sentry.LevelFatal)
					})
				}

				if err == nil {
					err = errors.Errorf("panic in job: %v", panicErr)
				}
			} else if err != nil {
				jobLog.WithError(err).Error("error while processing job")
				if hub != nil {
					hub.ConfigureScope(func(scope *sentry.Scope) {
						scope.SetLevel(sentry.LevelError)
					})
					hub.CaptureException(err)
				}
			}

			p.markJobStatus(span.Context(), job, err)
		}()
		defer span.Finish()

		jobLog.Trace("handling job")

		// Set err outright to make sentry reporting easier.
		err = handler.HandleConsumeJob(span.Context(), job.Input)
		return
	}
}
