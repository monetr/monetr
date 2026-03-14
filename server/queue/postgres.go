// Outline of how the postgres queue works:
//
//	           DATA FLOW — enqueue to execution
//
//	            queue.Enqueue()
//	                 │
//	           INSERT INTO jobs
//	           (status=pending, priority=unix_ts,
//	            signature=fnv32(args+ts))  ◄── ON CONFLICT: skip
//	                 │
//	           notifier.notify()  ──► notifier.channel ─────────────┐
//	                                                                │
//	┌────────────────────────────────────────────────────────────┐  │
//	│  jobConsumer wakes via notifier or ticker                  │ ◄┘
//	│       │                                                    │
//	│  SELECT job FOR UPDATE SKIP LOCKED                         │
//	│  WHERE status=pending AND priority<=now AND queue IN (...) │
//	│  UPDATE SET status=processing, started_at=now              │
//	│       │                                                    │
//	│  dispatch ──────────────────────────────────────────────►  │
//	└────────────────────────────────────────────────────────────┘
//	                                     │
//	                                worker receives
//	                                     │
//	                               executeJob()
//	                                     │
//	                           UPDATE job SET status=?
//	                           (completed / pending+retry / failed)
package queue

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sort"
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/storage"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
)

var (
	// workerSignal is a pointer to an empty struct that will get passed around
	// via channels. It means nothing on its own, and depends entirely on which
	// channel the signal is sent on.
	workerSignal struct{}
)

const (
	numberOfWorkers = 4
	maxAttempts     = 5
	attemptBackoff  = 10 * time.Second
)

const (
	postgresProcessorUninitialized = 0
	postgresProcessorRunning       = 1
	postgresProcessorStopped       = 2
)

var (
	_ Processor = &postgresProcessor{}
	_ Context   = &postgresContext{}
)

type postgresContext struct {
	ctx context.Context
	job models.Job
	*postgresProcessor
}

// RunInTransaction implements [Context].
// This function calls the provided callback with a modified version of the
// current context. The [Context.DB] function is now transactional within the
// callback. This way the changes made to data in the datastore within the
// callback is atomic. Other parts of the context however are not transactional.
// Such as email or billing. But [Enqueue] or [EnqueueAt] within the callback is
// transactional.
func (p *postgresContext) RunInTransaction(ctx context.Context, callback func(ctx Context) error) error {
	return p.DB().RunInTransaction(ctx, func(tx *pg.Tx) error {
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		// Clone the postgresProcessor and replace the DB with the transaction.
		// This must be a deep clone of the processor (not just the context) so
		// that replacing db does not mutate the original processor's db field
		// through the shared pointer.
		processorClone := *p.postgresProcessor
		processorClone.db = tx

		// Build a new context pointing at the cloned processor.
		clone := postgresContext{
			ctx:               span.Context(),
			postgresProcessor: &processorClone,
		}

		// Then call the callback using the cloned context
		return callback(&clone)
	})
}

func (p *postgresContext) Job() models.Job {
	return p.job
}

// Clock implements [Context].
func (p *postgresContext) Clock() clock.Clock {
	return p.clock
}

// Billing implements [Context].
func (p *postgresContext) Billing() billing.Billing {
	return p.billing
}

// DB implements [Context].
func (p *postgresContext) DB() pg.DBI {
	return p.db
}

// Deadline implements [Context].
// Subtle: this method shadows the method (Context).Deadline of postgresContext.Context.
func (p *postgresContext) Deadline() (deadline time.Time, ok bool) {
	return p.ctx.Deadline()
}

// Done implements [Context].
// Subtle: this method shadows the method (Context).Done of postgresContext.Context.
func (p *postgresContext) Done() <-chan struct{} {
	return p.ctx.Done()
}

// Email implements [Context].
func (p *postgresContext) Email() communication.EmailCommunication {
	return p.email
}

// Err implements [Context].
// Subtle: this method shadows the method (Context).Err of postgresContext.Context.
func (p *postgresContext) Err() error {
	return p.ctx.Err()
}

// KMS implements [Context].
func (p *postgresContext) KMS() secrets.KeyManagement {
	return p.kms
}

// Log implements [Context].
func (p *postgresContext) Log() *slog.Logger {
	return p.log
}

// Platypus implements [Context].
func (p *postgresContext) Platypus() platypus.Platypus {
	return p.plaidPlatypus
}

// Enqueuer implements [Context].
func (p *postgresContext) Enqueuer() Enqueuer {
	return p.postgresProcessor
}

// Publisher implements [Context].
func (p *postgresContext) Publisher() pubsub.Publisher {
	return p.publisher
}

// Storage implements [Context].
func (p *postgresContext) Storage() storage.Storage {
	return p.fileStorage
}

// Value implements [Context].
// Subtle: this method shadows the method (Context).Value of postgresContext.Context.
func (p *postgresContext) Value(key any) any {
	return p.ctx.Value(key)
}

type postgresProcessor struct {
	// state keeps track of what state the postgresProcessor is in. It is tracked
	// atomically against const values in order to make sure that certain actions
	// can only happen when the processor is in a certain state.
	state uint32

	// These values are provided in the [NewPostgresQueue] constructor, all other
	// fields are constructed as part of the [*postgresProcessor.Start] function.
	notifier      Notifier
	log           *slog.Logger
	clock         clock.Clock
	configuration config.Configuration
	db            pg.DBI
	publisher     pubsub.Publisher
	plaidPlatypus platypus.Platypus
	kms           secrets.KeyManagement
	fileStorage   storage.Storage
	billing       billing.Billing
	email         communication.EmailCommunication

	// jobQuery is a predetermined query that the postgresProcessor used in order
	// to retrieve jobs from the queue to be processed. This query is built when
	// the queue starts and is based on all of the registered jobs in the
	// processor. Jobs can't be registered after the processor starts so this
	// query is calculated right then so it can be re-used over and over again.
	jobQuery *pg.Query
	// queues represents an array of queue names used to build the jobQuery. This
	// array is appended as handlers are registered.
	queues []string
	// cronJobQueues keeps track of the cron jobs that need to be managed in the
	// cron table. These items are still appended to the queues array as they are
	// consumed from the cron table and enqueued immediately to the jobs table.
	cronJobQueues []string
	cronSchedules []struct {
		queue    string
		schedule string
		cron     cron.Schedule
		next     time.Time
	}
	// registeredJobs keeps track of the callback function for the actual job to
	// be executed per queue. A queue can only have a single job registered.
	registeredJobs map[string]InternalJobWrapper

	availableWorkers  chan struct{}
	shutdownConsumers []chan chan struct{}
	shutdownWorkers   []chan chan struct{}
	dispatch          chan *models.Job
}

// NewPostgresQueue creates a processor that is backed by a PostgreSQL database.
//
//	                  SYSTEM OVERVIEW — postgresProcessor
//
//	External callers
//	──────────────────────────────────────────────────────────────────
//	Enqueue() / EnqueueAt()
//	  │  INSERT job row (status=pending, priority=unix_ts)
//	  │  ON CONFLICT (signature) DO NOTHING  ← deduplication
//	  └─► notifier.notify()  ───────────────────────────────────────────────┐
//	                                                                        │
//	Channels (created in Start())                                           │
//	──────────────────────────────────────────────────────────────────      │
//	                                                                        │
//	  availableWorkers  chan struct{}  cap=4   (worker tokens)              │
//	  dispatch          chan *Job      cap=0   (unbuffered — must handoff)  │
//	  notifier.channel  chan struct{}  cap=N   ◄────────────────────────────┘
//	  shutdownConsumers []chan chan struct{}    (one per consumer goroutine)
//	  shutdownWorkers   []chan chan struct{}    (one per worker goroutine)
//
//	Goroutines (started in Start())
//	──────────────────────────────────────────────────────────────────
//
//	  jobConsumer  ×1  ──── reads availableWorkers, dispatch, notifier.channel
//	  cronConsumer ×1  ──── only if cron jobs are registered
//	  worker       ×4  ──── reads dispatch, writes availableWorkers
func NewPostgresQueue(
	ctx context.Context,
	notifier Notifier,
	clock clock.Clock,
	log *slog.Logger,
	configuration config.Configuration,
	db *pg.DB,
	publisher pubsub.Publisher,
	plaidPlatypus platypus.Platypus,
	kms secrets.KeyManagement,
	fileStorage storage.Storage,
	billing billing.Billing,
	email communication.EmailCommunication,
) Processor {
	return &postgresProcessor{
		notifier:       notifier,
		clock:          clock,
		log:            log,
		configuration:  configuration,
		db:             db,
		publisher:      publisher,
		plaidPlatypus:  plaidPlatypus,
		kms:            kms,
		fileStorage:    fileStorage,
		billing:        billing,
		email:          email,
		registeredJobs: map[string]InternalJobWrapper{},
	}
}

// EnqueueAt implements [Processor].
func (p *postgresProcessor) EnqueueAt(
	ctx context.Context,
	queue string,
	at time.Time,
	args any,
) error {
	span := sentry.StartSpan(ctx, "queue.publish")
	defer span.Finish()
	span.Description = queue
	span.SetTag("queue", queue)
	span.SetData("messaging.destination.name", queue)
	span.SetData("messaging.system", "postgresql")
	span.Data = map[string]any{
		"queue":     queue,
		"arguments": args,
	}

	crumbs.Debug(
		span.Context(),
		"Enqueueing job for background processing",
		map[string]any{
			"queue":     queue,
			"arguments": args,
		},
	)

	log := p.log.With("queue", queue)

	log.DebugContext(span.Context(), "enqueuing job to be run")

	encodedArgs, err := encodeArguments(args)
	if err != nil {
		return errors.Wrapf(err, "failed to encode arguments for job: %s", queue)
	}

	// TODO Make sure this truncation works how I think it does. Reasonably there
	// should be no duplicate jobs within 1 second of this job.
	timestamp := at.Truncate(time.Second)
	signature := jobSignature(timestamp, encodedArgs)
	traceId := span.ToSentryTrace()
	baggage := span.ToBaggage()
	job := models.Job{
		Queue:         queue,
		Signature:     signature,
		Priority:      uint64(timestamp.Unix()),
		Input:         string(encodedArgs),
		Output:        "",
		Status:        models.PendingJobStatus,
		SentryTraceId: &traceId,
		SentryBaggage: &baggage,
		Attempt:       1,
		CreatedAt:     timestamp,
		UpdatedAt:     timestamp,
		StartedAt:     nil,
		CompletedAt:   nil,
	}

	// Actually insert the job into the queue, but if the job fails to insert due
	// to a conflict on the signature column, then log a trace message and do
	// nothing.
	_, err = p.db.ModelContext(span.Context(), &job).Insert(&job)
	if err != nil {
		if pgErr, ok := err.(pg.Error); ok && pgErr.Field(67) == "23505" {
			// Do nothing. It is a duplicate enqueue.
			log.Log(
				span.Context(),
				logging.LevelTrace,
				"job has already been enqueued, it will not be enqueued again",
				"signature", signature,
			)
			return nil
		}
		return errors.Wrap(err, "failed to enqueue job")
	}

	log = p.log.With(
		"jobId", job.JobId,
		"signature", signature,
	)

	span.SetData("messaging.message.id", string(job.JobId))
	span.SetData("messaging.message.body.size", len(job.Input))

	log.DebugContext(span.Context(), "successfully enqueued job")

	// Notify the workers that there is a job available for consumption. This is
	// non-blocking in the sense that this wont block if there are no workers
	// available to consume the job right now.
	if err := p.notifier.notify(span.Context()); err != nil {
		log.DebugContext(span.Context(), "failed to notify workers of new jobs",
			"err", err,
		)
	}

	return nil
}

// Register implements [Processor].
func (p *postgresProcessor) Register(
	ctx context.Context,
	queue string,
	job InternalJobWrapper,
) error {
	if atomic.LoadUint32(&p.state) != postgresProcessorUninitialized {
		return errors.New("jobs cannot be added to the processor after it has been started or closed")
	}

	log := p.log.With("queue", queue)
	log.Log(ctx, logging.LevelTrace, "registering job handler")

	// Does not need to be wrapped in a mutex or anything because we only allow
	// the jobs to be registered before the processor has started. After it has
	// started this map must be READ ONLY!
	if _, ok := p.registeredJobs[queue]; ok {
		return errors.Errorf(
			"job handler has already been registered: %s",
			queue,
		)
	}

	p.registeredJobs[queue] = job
	p.queues = append(p.queues, queue)

	return nil
}

// RegisterCron implements [Processor].
func (p *postgresProcessor) RegisterCron(
	ctx context.Context,
	queue string,
	schedule string,
	job InternalJobWrapper,
) error {
	if atomic.LoadUint32(&p.state) != postgresProcessorUninitialized {
		return errors.New("jobs cannot be added to the processor after it has been started or closed")
	}

	cronSchedule, err := cron.Parse(schedule)
	if err != nil {
		return errors.Wrapf(err, "failed to parse cron schedule for job: %s - %s", queue, schedule)
	}

	log := p.log.With("queue", queue, "schedule", schedule)
	log.Log(ctx, logging.LevelTrace, "registering cron job handler")

	// Does not need to be wrapped in a mutex or anything because we only allow
	// the jobs to be registered before the processor has started. After it has
	// started this map must be READ ONLY!
	if _, ok := p.registeredJobs[queue]; ok {
		return errors.Errorf(
			"job handler has already been registered: %s",
			queue,
		)
	}

	p.registeredJobs[queue] = job
	p.queues = append(p.queues, queue)
	p.cronJobQueues = append(p.cronJobQueues, queue)
	p.cronSchedules = append(p.cronSchedules, struct {
		queue    string
		schedule string
		cron     cron.Schedule
		next     time.Time
	}{
		queue:    queue,
		schedule: schedule,
		cron:     cronSchedule,
	})

	return nil
}

// hydrateCronJobTable takes all of the registered cron jobs with this processor
// and assumes that these are the current canonical list of cron jobs that need
// to be processed. It will remove crons from the table that are no longer apart
// of that list, as well as add new cron jobs if necessary, and update existing
// cron jobs with new timings and schedules if they have changed.
func (p *postgresProcessor) hydrateCronJobTable() error {
	// If there are no cron jobs to hydrate then do nothing. We don't care if
	// there are some still in the table because we aren't going to do anything
	// with them.
	if len(p.cronJobQueues) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		{ // Clean up cron jobs that are no longer registered.
			// TODO If there are no cron job queues registered then this breaks, but
			// should it instead just remove all of the cron jobs?
			result, err := txn.ModelContext(ctx, new(models.CronJob)).
				WhereIn(`"queue" NOT IN (?)`, p.cronJobQueues).
				Delete()
			if err != nil {
				return errors.Wrap(err, "failed to clean up old cron jobs")
			} else if affected := result.RowsAffected(); affected > 0 {
				p.log.InfoContext(
					ctx,
					"removed outdated cron job(s) from postgres",
					"removed", affected,
				)
			} else {
				p.log.DebugContext(ctx, "no outdated cron jobs were removed")
			}
		}

		now := p.clock.Now().UTC()
		var crons []models.CronJob
		{ // Predetermine what all the cron rows should look like!
			for _, cronJob := range p.cronSchedules {
				crons = append(crons, models.CronJob{
					Queue:        cronJob.queue,
					CronSchedule: cronJob.schedule,
					NextRunAt:    cronJob.cron.Next(now),
				})
			}
		}

		{ // Upsert/merge the cron rows with the table
			result, err := txn.ModelContext(ctx, &crons).
				OnConflict(`("queue") DO UPDATE`).
				Set(`"cron_schedule" = EXCLUDED.cron_schedule`).
				// If a cron schedule is updated such that it should execute sooner,
				// then update the next run at to be that sooner timestamp. Otherwise
				// keep the current timestamp if it is sooner.
				Set(`"next_run_at" = least(EXCLUDED."next_run_at", "cron_job"."next_run_at")`).
				// But only update the cron job if the next run at or cron schedule
				// field would actually change.
				Where(`"cron_job"."next_run_at" != least(EXCLUDED."next_run_at", "cron_job"."next_run_at")`).
				WhereOr(`"cron_job"."cron_schedule" != EXCLUDED.cron_schedule`).
				Returning("NULL").
				Insert()
			if err != nil {
				return errors.Wrap(err, "failed to provision cron jobs")
			}

			if affected := result.RowsAffected(); affected == 0 {
				p.log.DebugContext(ctx, "no cron jobs required updating")
			} else {
				p.log.InfoContext(ctx, "updated cron jobs", "updated", affected)
			}
		}

		return nil
	})
}

func (p *postgresProcessor) jobContext(ctx context.Context, job *models.Job) Context {
	// Clone the postgresProcessor in memory
	clone := *p
	clone.log = clone.log.With(
		"queue", job.Queue,
		"jobId", job.JobId,
	)
	return &postgresContext{
		ctx:               ctx,
		job:               *job,
		postgresProcessor: &clone,
	}
}

func (p *postgresProcessor) WithTransaction(db pg.DBI) Enqueuer {
	clone := *p
	clone.db = db
	return &clone
}

func (p *postgresProcessor) Start() error {
	if !atomic.CompareAndSwapUint32(
		&p.state,
		postgresProcessorUninitialized,
		postgresProcessorRunning,
	) {
		return errors.New("job processor is either already started, or is in an invalid state")
	}

	if len(p.registeredJobs) == 0 {
		// Reset the state so start could be called again.
		atomic.StoreUint32(&p.state, postgresProcessorUninitialized)
		return errors.New("cannot start processor with no jobs registered")
	}

	if err := p.hydrateCronJobTable(); err != nil {
		return err
	}

	{ // Setup query used to consume jobs in the actual loop
		// This query is precalculated here to make it more efficient, it will only
		// ever consume jobs that we are aware of. If new job queues are added in
		// the future while this process is running, this server will never consume
		// those jobs.
		p.jobQuery = p.db.Model(new(models.Job)).
			Column("job_id").
			// Only get jobs that are pending.
			Where(`"status" = ?`, models.PendingJobStatus).
			// Only get jobs that have a priority that is now or in the past.
			Where(`"priority" <= extract(epoch from now() at time zone 'utc')::integer`).
			// Only consume jobs we recognize.
			WhereIn(`"queue" IN (?)`, p.queues).
			Order(`job_id ASC`).
			For(`UPDATE SKIP LOCKED`).
			Limit(1)
	}

	// Unbuffered channel so that way jobs must be consumed!
	p.dispatch = make(chan *models.Job)

	{
		// Channel should have exactly as many items as we have available threads.
		// When threads become available they will put the [p.availableWorkers] struct
		// into this channel.
		p.availableWorkers = make(chan struct{}, numberOfWorkers)
		p.shutdownWorkers = make([]chan chan struct{}, numberOfWorkers)
		for i := 0; i < numberOfWorkers; i++ {
			// TODO I don't think this channel should be buffered?
			p.shutdownWorkers[i] = make(chan chan struct{})
			go p.worker(p.shutdownWorkers[i])
		}
	}

	p.shutdownConsumers = make([]chan chan struct{}, 0, 2)
	{ // Always start the job consumer
		p.shutdownConsumers = append(p.shutdownConsumers, make(chan chan struct{}))
		go p.jobConsumer(p.shutdownConsumers[len(p.shutdownConsumers)-1])
	}

	// Only start the cron job consumer if we have cron jobs to consume
	if len(p.cronJobQueues) > 0 {
		p.shutdownConsumers = append(p.shutdownConsumers, make(chan chan struct{}))
		go p.cronConsumer(p.shutdownConsumers[len(p.shutdownConsumers)-1])
	}

	return nil
}

// Close will shutdown the queue processor. It will return nil if all of the
// threads for the processor are shut down gracefully. It will return an error
// if it times out shutting down. This also drains the workers gracefully and if
// there are any jobs in progress it will wait until they are complete before
// returning (within the timeout).
//
//	                  Close() — SHUTDOWN SEQUENCE
//
//	Consumers must stop BEFORE workers.
//	If workers stopped first, jobConsumer could block forever on
//	dispatch ← job with nobody on the other end.
//
//	STEP 1 — signal each consumer (jobConsumer, cronConsumer)
//
//	  for each shutdownConsumers[i]:
//	    shutdownConsumers[i] ◄── promises  (send the ack channel)
//
//	  for each shutdownConsumers[i]:
//	    <── promises  (wait for ack)
//
//	STEP 2 — signal each worker (×4)
//
//	  for each shutdownWorkers[i]:
//	    shutdownWorkers[i] ◄── promises
//
//	  for each shutdownWorkers[i]:
//	    <── promises
//
//	STEP 3 — close all channels
//	  close(shutdownConsumers[i]...)
//	  close(shutdownWorkers[i]...)
//	  close(dispatch)
//	  close(availableWorkers)
//
//	timeout: 15s for the entire sequence
func (p *postgresProcessor) Close() error {
	if !atomic.CompareAndSwapUint32(
		&p.state,
		postgresProcessorRunning,
		postgresProcessorStopped,
	) {
		return errors.New("job processor is either already closed, or is in an invalid state")
	}

	p.log.Info("shutting down job queue")

	timer := time.NewTimer(15 * time.Second)

	{ // Shutdown all of the consumer threads
		promises := make(chan struct{}, len(p.shutdownConsumers))
		for i := range p.shutdownConsumers {
			select {
			case p.shutdownConsumers[i] <- promises:
				p.log.Debug(fmt.Sprintf(
					"sent shutdown signal to consumer %d of %d",
					i+1, len(p.shutdownConsumers),
				))
				continue
			case <-timer.C:
				return errors.New("timed out sending shutdown signal to consumer")
			}
		}

		for i := range p.shutdownConsumers {
			select {
			case <-promises:
				p.log.Debug(fmt.Sprintf(
					"successfully shutdown consumer %d of %d",
					i+1, len(p.shutdownConsumers),
				))
				continue
			case <-timer.C:
				return errors.New("timed out waiting for consumers to drain")
			}
		}
		close(promises)
	}

	{ // Shutdown all the background threads
		// We create a channel that is the exact size of the number of threads in
		// the background. This way as each thread drains this channel should fill
		// up.
		promises := make(chan struct{}, len(p.shutdownWorkers))
		for i := range p.shutdownWorkers {
			select {
			case p.shutdownWorkers[i] <- promises:
				p.log.Debug(fmt.Sprintf(
					"sent shutdown signal to worker %d of %d",
					i+1, len(p.shutdownWorkers),
				))
				continue
			case <-timer.C:
				return errors.New("timed out sending shutdown signal")
			}
		}

		for i := range p.shutdownWorkers {
			select {
			case <-promises:
				p.log.Debug(fmt.Sprintf(
					"successfully shutdown worker %d of %d",
					i+1, len(p.shutdownWorkers),
				))
				continue
			case <-timer.C:
				return errors.New("timed out waiting for workers to drain")
			}
		}
		close(promises)
	}

	p.log.Info("job processor threads shutdown, cleaning up")

	for _, channel := range p.shutdownConsumers {
		close(channel)
	}
	for _, channel := range p.shutdownWorkers {
		close(channel)
	}
	close(p.dispatch)
	close(p.availableWorkers)

	p.log.Info("job processor shut down complete")
	return nil
}

func (p *postgresProcessor) consumeJobMaybe() (*models.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var job models.Job
	result, err := p.db.ModelContext(ctx, &job).
		Set(`"status" = ?`, models.ProcessingJobStatus).
		Set(`"started_at" = ?`, p.clock.Now()).
		Where(`"job_id" = (?)`, p.jobQuery).
		Returning("*; /* NO LOG */").
		Update(&job)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to consume job")
	}

	if result.RowsAffected() == 0 {
		// Do nothing, there either isn't a job or we didn't get one.
		return nil, nil
	}

	p.log.Log(ctx, logging.LevelTrace, "found job",
		"jobId", job.JobId,
		"queue", job.Queue,
	)

	return &job, nil
}

// consumeCronMaybe takes the queue and the next timestamp that that cron job
// should recur and attempts to consume it. It does this by updating the row for
// that cron job after querying the row from the table with a FOR UPDATE SKIP
// LOCKED query. This way even if there are multiple cron job workers
// (different server instances) only one will ever consume the cron job itself.
func (p *postgresProcessor) consumeCronMaybe(
	queue string,
	next time.Time,
) (*models.CronJob, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	subQuery := p.db.ModelContext(ctx, new(models.CronJob)).
		Column("queue").
		Where(`"queue" = ?`, queue).
		Where(`"next_run_at" < ?`, next).
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

// cronConsumer is the loop that maintains cron jobs. This loop runs on all
// servers, however the PostgreSQL database is used to ensure that the cron job
// is only consumed a single time.
//
//	                     cronConsumer LOOP
//
//	BOOTSTRAP (once):
//	  for each cronSchedule:
//	    schedule.next = cron.Next(now)
//
//	┌──────────────────────────────────────────────────────────────────────┐
//	│                                                                      │
//	│  sort cronSchedules by .next  (soonest first)                        │
//	│  nextJob = cronSchedules[0]                                          │
//	│  next    = nextJob.next                                              │
//	│  advance: cronSchedules[0].next = nextJob.cron.Next(next)            │
//	│                                                                      │
//	│  sleep = next.Sub(now)                                               │
//	│  timer = time.NewTimer(sleep)                                        │
//	│                                                                      │
//	│         ┌──────────────┐                                             │
//	│         │    select    │                                             │
//	│         └──────────────┘                                             │
//	│              /        \                                              │
//	│        shutdown       timer.C                                        │
//	│           │              │                                           │
//	│         return     consumeCronMaybe(queue, cron.Next(next))          │
//	│                          │                                           │
//	│                    ┌─────┴──────┐                                    │
//	│                  nil        consumed                                 │
//	│                    │              │                                  │
//	│                 continue    enqueueAt(queue, next, nil)              │
//	│                             (regular job row, deduped by sig)        │
//	│                             continue ────────────────────────────────┘
//	└──────────────────────────────────────────────────────────────────────┘
//
//	consumeCronMaybe:
//	  SELECT queue FOR UPDATE SKIP LOCKED
//	    WHERE queue = ? AND next_run_at < next
//	  UPDATE cron_job
//	    SET last_run_at = next_run_at, next_run_at = next
//	  → only one server in a cluster wins the race
func (p *postgresProcessor) cronConsumer(shutdown chan chan struct{}) {
	{ // Bootstrap the next timestamp before we do anything. We'll maintain this
		now := p.clock.Now().UTC()
		for i := range p.cronSchedules {
			p.cronSchedules[i].next = p.cronSchedules[i].cron.Next(now)
		}
	}

	for {
		now := p.clock.Now().UTC()
		// Sort the crons by their next time a cron job happens.
		sort.Slice(p.cronSchedules, func(i, j int) bool {
			return p.cronSchedules[i].next.Before(p.cronSchedules[j].next)
		})
		nextJob := p.cronSchedules[0]
		next := nextJob.next
		// Move the next timestamp forward, this way the next loop we are no loger
		// looking at this specific cron job. Instead we'll sort and get the next
		// most recent cron job.
		p.cronSchedules[0].next = nextJob.cron.Next(next)
		p.log.Log(
			context.Background(),
			slog.LevelDebug,
			"staged next cron job to be run",
			"queue", nextJob.queue,
			"next", next,
			"now", now,
		)

		// TODO What happens if sleep is negative or 0
		// How long do we need to wait for this cron job?
		sleep := next.Sub(now)

		// Create a timer.
		timer := time.NewTimer(sleep)
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down cron job consumer")
			timer.Stop()
			promise <- workerSignal
			return
		case <-timer.C:
			// Bump the cron we just did. But use a slightly more future timestamp.
			// This is to fix a bug where sometimes the cron library seems to be
			// rounding down? Resulting in a `nextTimestamp` that is slightly in the
			// past.
			log := p.log.With("queue", nextJob.queue)
			consumedCronJob, err := p.consumeCronMaybe(
				nextJob.queue,
				// Bump the cron to the next time it would be consumed
				nextJob.cron.Next(next),
			)
			if err != nil {
				log.Error("failed to consume cron job", "err", err)
				continue
			}
			if consumedCronJob == nil {
				log.Log(context.Background(), logging.LevelTrace, "did not consume cron job")
				continue
			}

			// If we actually did consume the cron job then log it
			log.Log(context.Background(), logging.LevelTrace, "consumed cron job")

			if err := p.EnqueueAt(
				context.Background(),
				nextJob.queue,
				// An accidental other form of deduplication. Cron jobs are always added
				// to the processing queue with their "next timestamp". This way the
				// timestamp is deterministic for them.
				next,
				nil,
			); err != nil {
				log.Error("failed to enqueue cron for execution", "err", err)
			}
		}
	}
}

// jobConsumer is the loop that takes jobs from the queue and if there are
// threads available, will dispatch the jobs to the worker threads.
//
//	                      jobConsumer LOOP
//
//	ticker := time.NewTicker(10s)    backoff: 1..6  (×10s = 10s..60s)
//
//	┌──────────────────────────────────────────────────────────────────────┐
//	│  PHASE 1 — wait for an available worker token                        │
//	│                                                                      │
//	│         ┌──────────────┐                                             │
//	│         │    select    │                                             │
//	│         └──────────────┘                                             │
//	│          /      |       \          \                                 │
//	│  shutdown  ticker.C  notifier   availableWorkers                     │
//	│     │        │           │          │                                │
//	│   return   log +       log +       claim token                       │
//	│            continue    continue    ──────────────────────────────►   │
//	│                                                                      │
//	│  CONSUME — consumeJobMaybe()  (SELECT ... FOR UPDATE SKIP LOCKED)    │
//	│                                                                      │
//	│   err != nil ──► return token to availableWorkers, log error         │
//	│   job == nil ──► return token to availableWorkers                    │
//	│                  backoff++ (if < max), reset ticker                  │
//	│   job != nil ──► dispatch ◄──── send *Job to worker                  │
//	│                  reset backoff=1, reset ticker                       │
//	│                                                                      │
//	│  PHASE 2 — wait for next trigger                                     │
//	│                                                                      │
//	│         ┌──────────────┐                                             │
//	│         │    select    │                                             │
//	│         └──────────────┘                                             │
//	│          /      |       \                                            │
//	│  shutdown  ticker.C  notifier                                        │
//	│     │        │           │                                           │
//	│   return  continue    continue ──────────────────────────────────────┘
//	└──────────────────────────────────────────────────────────────────────┘
func (p *postgresProcessor) jobConsumer(shutdown chan chan struct{}) {
	// maxBackoff is 60 seconds
	var maxBackoff time.Duration = 6
	var backoff time.Duration = 1
	baseInterval := 10 * time.Second
	ticker := time.NewTicker(baseInterval)
	for {
		// Because this is at the beginning of the loop, the only time this blocks
		// is when there are no available threads. The ticker will only fire here if
		// there are no available threads, which will simply log a message and reset
		// the loop.
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down job consumer")
			ticker.Stop()
			promise <- workerSignal
			return
		case <-ticker.C:
			p.log.Debug("job processor currently has no available worker threads")
			continue
		case <-p.notifier.channel():
			p.log.Debug("received job notification, but job processor currently has no available worker threads")
			continue
		case <-p.availableWorkers:
			// This receive blocks if there are no available threads.
		}

		// Once we have an available thread, try to consume a job.
		job, err := p.consumeJobMaybe()
		if err != nil {
			// If we experienced an error trying to pull a job from the queue then we
			// need to return an available thread to the channel. This way if the
			// database server is failing for just a moment we don't exhaust our
			// available threads.
			p.availableWorkers <- workerSignal
			p.log.Error("failed to consume job", "err", err)
		} else if job != nil {
			p.log.Debug("successfully consumed job, dispatching to worker thread",
				"jobId", job.JobId,
				"queue", job.Queue,
			)
			p.dispatch <- job
			// If we successfully consume a job then reset the backoff
			backoff = 1
			ticker.Reset(baseInterval)
		} else {
			// If we did not retrieve a job at all then we need to put our "hold" on
			// an available thread back in the channel this way a thread can still be
			// consumed on the next loop.
			p.availableWorkers <- workerSignal
			// If we haven't consumed a job in a while then backoff how often we poll
			// the database. We get notifications when we know a job is immediately
			// available so this should be fine.
			if backoff+1 <= maxBackoff {
				backoff++
				interval := backoff * baseInterval
				ticker.Reset(interval)
				p.log.Debug("no jobs consumed in a while, backing off",
					"interval", interval.String(),
				)
			}
		}

		// This is the main area that blocks. After we have consumed a job there
		// will always be available threads. So in this one we are ALWAYS blocking
		// on the ticker or the shutdown signal. When the ticker fires we will
		// proceed to the top of this loop. Where once again, we will only block if
		// there are no available threads.
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down job consumer")
			promise <- workerSignal
			ticker.Stop()
			return
		case <-p.notifier.channel():
			p.log.Debug("received job notification")
			continue
		case <-ticker.C:
			continue
		}
	}
}

// worker is the loop that runs based on the number of threads we want to have
// on each server. Each worker can execute one job at a time. When a worker is
// available to execute jobs it will send a signal on
// [*postgresProcessor.availableWorkers].
//
//	                        worker LOOP  (×4)
//
//	┌──────────────────────────────────────────────────────────────────────┐
//	│                                                                      │
//	│  availableWorkers ◄── send workerSignal  (token: "I am free")        │
//	│                                                                      │
//	│         ┌──────────────┐                                             │
//	│         │    select    │                                             │
//	│         └──────────────┘                                             │
//	│              /        \                                              │
//	│        shutdown      dispatch                                        │
//	│           │              │                                           │
//	│         return      executeJob(job)                                  │
//	│                          │                                           │
//	│              ┌───────────┴───────────┐                               │
//	│         shutdown check (select/default)                              │
//	│              │                   │                                   │
//	│           return              continue ──────────────────────────────┘
//	└──────────────────────────────────────────────────────────────────────┘
func (p *postgresProcessor) worker(shutdown chan chan struct{}) {
	for {
		// By adding an available thread to this channel, we tell the processor that
		// there are threads that can perform work. We do this at the beginning of
		// the loop so that way we always return to a state where we can perform
		// work instead of getting deadlocked.
		p.availableWorkers <- workerSignal

		// We block on channel reads for both the shutdown channel and the dispatch
		// channel. This way the thread for this worker is essentially dead until we
		// are either told to stop working entirely, or until we are told to perform
		// the work for a job.
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down worker thread")
			promise <- workerSignal
			return
		case job := <-p.dispatch:
			// Execute the job in a wrapper, this way if the job panics or anything we
			// have some safety.
			// TODO Maybe wrap this in a go routine with a channel blocking it? That
			// way we can notify if a job isn't terminating properly on timeout.
			p.executeJob(job)
		}

		// After we have executed the job then we want to check the shutdown signal
		// again. Because if we have shutdown then sending on the availableWorkers
		// channel will block or fail depending on timing. So we need to do this
		// first.
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down worker thread")
			promise <- workerSignal
			return
		default:
			// If we aren't shutting down then proceed!
			continue
		}
	}
}

// executeJob is a wrapper around the actual execution. The error handling and
// retry logic for every job is implemented here, as well as tracing and more.
//
//	                     executeJob — JOB OUTCOMES
//
//	 executor(jobContext, args)
//	       │
//	  ┌────┴────────────────────────────────────────┐
//	  │                                             │
//	 panic                                        return
//	  │                                       /         \
//	mark FAILED                            err            nil
//	(no retry)                              │               │
//	                               attempt < maxAttempts   mark COMPLETED
//	                                    /        \
//	                                retry       exhausted
//	                                  │              │
//	                            mark PENDING     mark FAILED
//	                            priority += backoff×attempt
//	                            (10s, 20s, 30s, 40s, 50s)
func (p *postgresProcessor) executeJob(job *models.Job) {
	executor, ok := p.registeredJobs[job.Queue]
	if !ok {
		panic(fmt.Sprintf(
			"could not execute job with queue name: %s no handler registered",
			job.Queue,
		))
	}

	log := p.log.With(
		"jobId", job.JobId,
		"queue", job.Queue,
	)
	log.Info("processing job")

	// Execute the job with a timeout.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	highContext := sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())
	options := []sentry.SpanOption{
		sentry.WithTransactionName(job.Queue),
	}
	if job.SentryTraceId != nil && job.SentryBaggage != nil {
		options = append(options, sentry.ContinueFromHeaders(
			*job.SentryTraceId,
			*job.SentryBaggage,
		))
	}
	span := sentry.StartSpan(
		highContext,
		"queue.process",
		options...,
	)
	span.Description = job.Queue
	span.SetData("input", job.Input)
	span.SetData("messaging.message.id", string(job.JobId))
	span.SetData("messaging.destination.name", string(job.Queue))
	span.SetData("messaging.message.body.size", len(job.Input))
	span.SetData("messaging.message.receive.latency", time.Now().Sub(job.CreatedAt).Milliseconds())
	span.SetData("messaging.system", "postgresql")
	// For now, sample all background jobs
	span.Sampled = sentry.SampledTrue
	hub := sentry.GetHubFromContext(span.Context())
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("queue", job.Queue)
	})

	var err error
	defer func() {
		if panicErr := recover(); panicErr != nil {
			log.ErrorContext(
				span.Context(),
				"panic while processing job",
				"err", panicErr,
				"stackTrace", string(debug.Stack()),
			)
			if hub != nil {
				hub.RecoverWithContext(span.Context(), panicErr)
				hub.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetLevel(sentry.LevelFatal)
				})
			}

			if err == nil {
				err = errors.Errorf("panic in job: %v", panicErr)
			}
			span.Status = sentry.SpanStatusInternalError

			// If we panic during the job then just mark the job as failed and don't
			// retry.
			now := p.clock.Now()
			if _, err := p.db.Model(job).
				Set(`"status" = ?`, models.FailedJobStatus).
				Set(`"completed_at" = ?`, now).
				Set(`"updated_at" = ?`, now).
				WherePK().
				Update(); err != nil {
				log.ErrorContext(span.Context(), "failed to mark job as failed", "err", err)
			}
		} else if err != nil {
			log.ErrorContext(span.Context(), "error while processing job", "err", err)
			if hub != nil {
				hub.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetLevel(sentry.LevelError)
				})
				hub.CaptureException(err)
			}
			span.Status = sentry.SpanStatusInternalError

			// If we get a regular error back and we have not exhausted our retries
			// then we can bump the job again.
			if job.Attempt < maxAttempts {
				log.WarnContext(span.Context(), "job has attempts left and will be retried",
					"err", err,
					"attempt", job.Attempt,
				)
				now := p.clock.Now()
				if _, err := p.db.Model(job).
					Set(`"attempt" = ?`, job.Attempt+1).
					Set(`"priority" = ?`, now.Add(attemptBackoff*time.Duration(job.Attempt)).Unix()).
					Set(`"status" = ?`, models.PendingJobStatus).
					Set(`"started_at" = NULL`).
					Set(`"updated_at" = ?`, now).
					WherePK().
					Update(); err != nil {
					log.ErrorContext(span.Context(), "failed to bump job for retry", "err", err)
				}
			} else {
				// If we have exhausted our retries though then move the job to a failed
				// status.
				log.WarnContext(span.Context(), "job has no remaining attempts, marking as failed",
					"err", err,
					"attempt", job.Attempt,
				)
				now := p.clock.Now()
				if _, err := p.db.Model(job).
					Set(`"status" = ?`, models.FailedJobStatus).
					Set(`"completed_at" = ?`, now).
					Set(`"updated_at" = ?`, now).
					WherePK().
					Update(); err != nil {
					log.ErrorContext(span.Context(), "failed to mark job as failed", "err", err)
				}
			}
		} else {
			// If the job succeeded then mark the job as completed and we don't need
			// to do anything else.
			now := p.clock.Now()
			if _, err := p.db.Model(job).
				Set(`"status" = ?`, models.CompletedJobStatus).
				Set(`"completed_at" = ?`, now).
				Set(`"updated_at" = ?`, now).
				WherePK().
				Update(); err != nil {
				log.ErrorContext(span.Context(), "failed to mark job as complete", "err", err)
			}

			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelInfo)
			})
			span.Status = sentry.SpanStatusOK
		}

		span.Finish()
	}()

	innerSpan := sentry.StartSpan(span.Context(), "job.exec")
	defer innerSpan.Finish()

	if err = executor(
		p.jobContext(innerSpan.Context(), job),
		[]byte(job.Input),
	); err != nil {
		log.Error("failed to execute job", "err", err)
	}
}
