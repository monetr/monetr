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
)

const (
	postgresProcessorUninitialized = 0
	postgresProcessorRunning       = 1
	postgresProcessorStopped       = 2
)

var (
	_ Processor = &postgresProcessor{}
)

type postgresProcessor struct {
	// state keeps track of what state the postgresProcessor is in. It is tracked
	// atomically against const values in order to make sure that certain actions
	// can only happen when the processor is in a certain state.
	state uint32

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
	}
	// registeredJobs keeps track of the callback function for the actual job to
	// be executed per queue. A queue can only have a single job registered.
	registeredJobs map[string]internalJobWrapper

	availableThreads chan struct{}
	shutdownThreads  []chan chan struct{}
	dispatch         chan *models.Job
}

func NewPostgresQueue(
	ctx context.Context,
	log *slog.Logger,
	clock clock.Clock,
	configuration config.Configuration,
	db *pg.DB,
	publisher pubsub.Publisher,
	plaidPlatypus platypus.Platypus,
	kms secrets.KeyManagement,
	fileStorage storage.Storage,
	billing billing.Billing,
	email communication.EmailCommunication,
) Processor {
	return &postgresProcessor{}
}

// enqueue implements [Processor].
func (p *postgresProcessor) enqueue(
	ctx context.Context,
	queue string,
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
	timestamp := p.clock.Now().UTC().Truncate(time.Second)
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
		Attempt:       0,
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

	span.SetData("messaging.message.id", string(job.JobId))
	span.SetData("messaging.message.body.size", len(job.Input))

	log.DebugContext(span.Context(), "successfully enqueued job",
		"jobId", job.JobId,
		"signature", signature,
	)

	// TODO Send notification for the job?

	panic("unimplemented")
}

// register implements [Processor].
func (p *postgresProcessor) register(
	ctx context.Context,
	queue string,
	job internalJobWrapper,
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

// registerCron implements [Processor].
func (p *postgresProcessor) registerCron(
	ctx context.Context,
	queue string,
	schedule string,
	job internalJobWrapper,
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.db.RunInTransaction(ctx, func(txn *pg.Tx) error {
		{ // Clean up cron jobs that are no longer registered.
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
			result, err := txn.ModelContext(ctx, crons).
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

func (p *postgresProcessor) beginTransaction() (Processor, error) {
	txn, err := p.db.Begin()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Clone the postgresProcessor in memory
	clone := *p
	// Overwrite the database item on the COPY of the processor.
	clone.db = txn
	return &clone, nil
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

	// Channel should have exactly as many items as we have available threads.
	// When threads become available they will put the [availableThread] struct
	// into this channel.
	p.availableThreads = make(chan struct{}, numberOfWorkers)
	p.shutdownThreads = make([]chan chan struct{}, numberOfWorkers+2)
	for i := 0; i < numberOfWorkers; i++ {
		// TODO I don't think this channel should be buffered?
		p.shutdownThreads[i] = make(chan chan struct{})
		go p.worker(p.shutdownThreads[i])
	}
	// use the last item as a way to shutdown the consumer
	p.shutdownThreads[numberOfWorkers] = make(chan chan struct{})
	go p.cronConsumer(p.shutdownThreads[numberOfWorkers])
	p.shutdownThreads[numberOfWorkers+1] = make(chan chan struct{})
	go p.jobConsumer(p.shutdownThreads[numberOfWorkers+1])

	return nil
}

func (p *postgresProcessor) Close() error {
	if !atomic.CompareAndSwapUint32(
		&p.state,
		postgresProcessorRunning,
		postgresProcessorStopped,
	) {
		return errors.New("job processor is either already closed, or is in an invalid state")
	}

	timer := time.NewTimer(15 * time.Second)

	{ // Shutdown all the background threads
		// We create a channel that is the exact size of the number of threads in
		// the background. This way as each thread drains this channel should fill
		// up.
		promises := make(chan struct{}, len(p.shutdownThreads))
		for i := range p.shutdownThreads {
			select {
			case p.shutdownThreads[i] <- promises:
				continue
			case <-timer.C:
				return errors.New("timed out sending shutdown signal")
			}
		}

		for i := 0; i < len(promises); i++ {
			select {
			case <-promises:
				p.log.Log(
					context.Background(),
					logging.LevelTrace,
					"thread successfully drained",
				)
				continue
			case <-timer.C:
				return errors.New("timed out waiting for threads to drain")
			}
		}
		close(promises)
	}

	p.log.Info("job processor threads shutdown, cleaning up")

	for _, channel := range p.shutdownThreads {
		close(channel)
	}
	close(p.dispatch)
	close(p.availableThreads)

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

func (p *postgresProcessor) cronConsumer(shutdown chan chan struct{}) {
	for {
		now := p.clock.Now().UTC()
		// Sort the crons by their next time a cron job happens.
		sort.Slice(p.cronSchedules, func(i, j int) bool {
			return p.cronSchedules[i].cron.Next(now).Before(p.cronSchedules[j].cron.Next(now))
		})
		nextJob := p.cronSchedules[0]
		next := nextJob.cron.Next(now)
		p.log.Log(
			context.Background(),
			logging.LevelTrace,
			"staged next cron job to be run",
			"queue", nextJob.queue,
			"next", next,
			"now", p.clock.Now(),
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
			consumedCronJob, err := p.consumeCronMaybe(nextJob.queue, next)
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

			if err := p.enqueue(context.Background(), nextJob.queue, nil); err != nil {
				log.Error("failed to enqueue cron for execution", "err", err)
			}
		}
	}
}

func (p *postgresProcessor) jobConsumer(shutdown chan chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
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
		case <-p.availableThreads:
			// This receive blocks if there are no available threads.
		}

		// Once we have an available thread, try to consume a job.
		job, err := p.consumeJobMaybe()
		if err != nil {
			// If we experienced an error trying to pull a job from the queue then we
			// need to return an available thread to the channel. This way if the
			// database server is failing for just a moment we don't exhaust our
			// available threads.
			p.availableThreads <- workerSignal
			p.log.Error("failed to consume job", "err", err)
		} else if job != nil {
			p.log.Debug("successfully consumed job, dispatching to worker thread",
				"jobId", job.JobId,
				"queue", job.Queue,
			)
			p.dispatch <- job
		} else {
			// If we did not retrieve a job at all then we need to put our "hold" on
			// an available thread back in the channel this way a thread can still be
			// consumed on the next loop.
			p.availableThreads <- workerSignal
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
		case <-ticker.C:
			continue
		}
	}
}

func (p *postgresProcessor) worker(shutdown chan chan struct{}) {
	for {
		// By adding an available thread to this channel, we tell the processor that
		// there are threads that can perform work. We do this at the beginning of
		// the loop so that way we always return to a state where we can perform
		// work instead of getting deadlocked.
		p.availableThreads <- workerSignal

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
	}
}

// executeJob is a wrapper around the actual execution. The error handling and
// retry logic for every job is implemented here, as well as tracing and more.
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
	span.SetData("messaging.message.receive.latency", p.clock.Now().Sub(job.CreatedAt).Milliseconds())
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
			log.ErrorContext(span.Context(), fmt.Sprintf("panic while processing job\n%+v\n%s", panicErr, string(debug.Stack())))
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
		} else if err != nil {
			log.ErrorContext(span.Context(), "error while processing job", "err", err)
			if hub != nil {
				hub.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetLevel(sentry.LevelError)
				})
				hub.CaptureException(err)
			}
			span.Status = sentry.SpanStatusInternalError
		} else {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelInfo)
			})
			span.Status = sentry.SpanStatusOK
		}
		// TODO Mark the job status as failed? Retry logic?
		span.Finish()
	}()

	if err := executor(nil, []byte(job.Input)); err != nil {
		// TODO Implement retry logic here?
		log.Error("failed to execute job", "err", err)
	}
}
