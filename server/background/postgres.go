package background

import (
	"context"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"runtime/debug"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

var (
	consumerSignal = struct{}{}
)

const (
	numberOfPostgresQueueWorkers = 4
	jobTimeoutSeconds            = 120
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

// Deprecated: Use EnqueueJobTxn instead.
func (p *postgresJobEnqueuer) EnqueueJob(ctx context.Context, queue string, arguments interface{}) error {
	return p.EnqueueJobTxn(ctx, p.db, queue, arguments)
}

func (p *postgresJobEnqueuer) EnqueueJobTxn(ctx context.Context, txn pg.DBI, queue string, arguments interface{}) error {
	span := sentry.StartSpan(ctx, "queue.publish")
	defer span.Finish()
	span.Description = queue
	span.SetTag("queue", queue)
	span.SetData("messaging.destination.name", queue)
	span.SetData("messaging.system", "postgresql")
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

	traceId := span.ToSentryTrace()
	baggage := span.ToBaggage()
	job := models.Job{
		Queue:         queue,
		Signature:     signature,
		Priority:      uint64(timestamp.Unix()),
		Input:         string(encodedArguments),
		Output:        "",
		Status:        models.PendingJobStatus,
		SentryTraceId: &traceId,
		SentryBaggage: &baggage,
		CreatedAt:     timestamp,
		UpdatedAt:     timestamp,
		StartedAt:     nil,
		CompletedAt:   nil,
	}
	_, err = txn.ModelContext(span.Context(), &job).Insert(&job)
	if err != nil {
		if pgErr, ok := err.(pg.Error); ok && pgErr.Field(67) == "23505" {
			// Do nothing. It is a duplicate enqueue.
			log.WithField("signature", signature).Trace("job has already been enqueued, it will not be enqueued again")
			return nil
		}
		return errors.Wrap(err, "failed to enqueue job for postgres")
	}
	span.SetData("messaging.message.id", string(job.JobId))
	span.SetData("messaging.message.body.size", len(job.Input))

	log.WithFields(logrus.Fields{
		"jobId":     job.JobId,
		"signature": signature,
	}).Debug("successfully enqueued job")

	return nil
}

type postgresJobFunction func(ctx context.Context, job *models.Job) error

type postgresJobProcessor struct {
	state                   uint32
	availableThreads        chan struct{}
	shutdownConsumerThreads []chan chan struct{}
	shutdownWorkerThreads   []chan chan struct{}
	dispatch                chan *models.Job
	trigger                 chan struct{}
	cronJobQueues           []ScheduledJobHandler
	queues                  []string
	registeredJobs          map[string]postgresJobFunction
	jobQuery                *pg.Query
	clock                   clock.Clock
	log                     *logrus.Entry
	db                      pg.DBI
	enqueuer                JobEnqueuer
	marshal                 JobMarshaller
}

func NewPostgresJobProcessor(
	log *logrus.Entry,
	clock clock.Clock,
	db pg.DBI,
	enqueuer JobEnqueuer,
) *postgresJobProcessor {
	return &postgresJobProcessor{
		shutdownConsumerThreads: []chan chan struct{}{},
		dispatch:                make(chan *models.Job),
		trigger:                 make(chan struct{}),
		cronJobQueues:           []ScheduledJobHandler{},
		queues:                  []string{},
		registeredJobs:          map[string]postgresJobFunction{},
		clock:                   clock,
		log:                     log,
		db:                      db,
		enqueuer:                enqueuer,
		marshal:                 DefaultJobMarshaller,
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

	if scheduledJob, ok := handler.(ScheduledJobHandler); ok {
		schedule := scheduledJob.DefaultSchedule()
		log.WithField("schedule", schedule).
			Trace("job will be run on a schedule automatically")
		p.cronJobQueues = append(p.cronJobQueues, scheduledJob)
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
		// Only get jobs that are pending.
		Where(`"status" = ?`, models.PendingJobStatus).
		// Only get jobs that have a priority that is now or in the past.
		Where(`"priority" <= extract(epoch from now() at time zone 'utc')::integer`).
		// Only consume jobs we recognize.
		WhereIn(`"queue" IN (?)`, p.queues).
		Order(`created_at ASC`).
		For(`UPDATE SKIP LOCKED`).
		Limit(1)

	numberOfWorkers := numberOfPostgresQueueWorkers
	numberOfConsumerThreads := 1 // Minimum of one for consuming jobs

	// If we are also consuming crons then we need an additional supporting
	// thread.
	if len(p.cronJobQueues) > 0 {
		numberOfConsumerThreads += 1
	}

	{ // Worker threads that actually perform the jobs
		p.availableThreads = make(chan struct{}, numberOfWorkers)
		p.shutdownWorkerThreads = make([]chan chan struct{}, numberOfWorkers)
		// Start the worker threads.
		for i := 0; i < numberOfWorkers; i++ {
			p.shutdownWorkerThreads[i] = make(chan chan struct{}, 1)
			go p.worker(p.shutdownWorkerThreads[i])
		}
	}

	{ // Supporting threads like job and cron consumers
		p.shutdownConsumerThreads = make([]chan chan struct{}, numberOfConsumerThreads)

		// Start the consumer thread.
		p.shutdownConsumerThreads[0] = make(chan chan struct{})
		go p.backgroundConsumer(p.shutdownConsumerThreads[0])

		// If there are any cron jobs registered then start the cron consumer.
		if len(p.cronJobQueues) > 0 {
			p.shutdownConsumerThreads[1] = make(chan chan struct{})
			go p.cronConsumer(p.shutdownConsumerThreads[1])
		}
	}

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
				// If a cron schedule is updated such that it should execute sooner,
				// then update the next run at to be that sooner timestamp. Otherwise
				// keep the current timestamp if it is sooner.
				Set(`"next_run_at" = least(EXCLUDED."next_run_at", ?)`, nextRunAt).
				// But only update the cron job if the next run at or cron schedule
				// field would actually change.
				Where(`"cron_job"."next_run_at" != least("cron_job"."next_run_at", ?)`, nextRunAt).
				WhereOr(`"cron_job"."cron_schedule" != ?`, cronJob.DefaultSchedule()).
				Returning("NULL").
				Insert()
			if err != nil {
				return errors.Wrapf(err, "failed to provision cron job: %s", cronJob)
			}

			if result.RowsAffected() == 0 {
				log.Debug("cron job already exists and did not need updating")
			} else {
				log.Info("cron job was updated")
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

	timer := time.NewTimer(15 * time.Second)

	p.log.Info("shutting down postgresql job processor")

	{ // Shutdown the consumers first
		// Create a channel buffer with the number of messages we need to send to
		// all the consumer threads.
		consumerShutdownChannel := make(chan struct{}, len(p.shutdownConsumerThreads))
		p.log.Debugf("shutting down %d postgresql job consumers", len(p.shutdownConsumerThreads))

		// Then send the shutdown channel to each consumer thread as a "promise".
		for i := range p.shutdownConsumerThreads {
			select {
			case p.shutdownConsumerThreads[i] <- consumerShutdownChannel:
				continue
			case <-timer.C:
				return errors.New("timed out sending shutdown signal to consumers")
			}
		}

		// Then wait for all consumers to be completely shutdown
		for i := 0; i < len(p.shutdownConsumerThreads); i++ {
			select {
			case <-consumerShutdownChannel:
				p.log.Trace("consumer successfully drained")
				continue
			case <-timer.C:
				return errors.New("timed out waiting for consumers to drain")
			}
		}
		close(consumerShutdownChannel)
	}

	{ // Then shutdown the workers
		// Create a channel buffer with the number of messages we need to send to
		// all the worker threads.
		workerShutdownChannel := make(chan struct{}, len(p.shutdownWorkerThreads))
		p.log.Debugf("shutting down %d postgresql job workers", len(p.shutdownWorkerThreads))

		// Then send the shutdown channel to each consumer thread as a "promise".
		for i := range p.shutdownWorkerThreads {
			select {
			case p.shutdownWorkerThreads[i] <- workerShutdownChannel:
				continue
			case <-timer.C:
				return errors.New("timed out sending shutdown signal to workers")
			}
		}

		// Then wait for all consumers to be completely shutdown
		for i := 0; i < len(p.shutdownWorkerThreads); i++ {
			select {
			case <-workerShutdownChannel:
				p.log.Trace("worker successfully drained")
				continue
			case <-timer.C:
				return errors.New("timed out waiting for workers to drain")
			}
		}
		close(workerShutdownChannel)
	}

	p.log.Info("postgresql job processor threads shutdown, cleaning up")

	for _, channel := range p.shutdownConsumerThreads {
		close(channel)
	}
	for _, channel := range p.shutdownWorkerThreads {
		close(channel)
	}
	close(p.trigger)
	close(p.dispatch)
	close(p.availableThreads)

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
			promise <- consumerSignal
			ticker.Stop()
			return
		case <-ticker.C:
			p.log.Trace("background consumer currently has no available worker threads")
			continue
		case <-p.trigger:
			p.log.Trace("received trigger notification, but background consumer has no available worker threads")
			continue
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
			// If we did not retrieve a job at all then we need to put our "hold" on
			// an available thread back in the channel this way a thread can still be
			// consumed on the next loop.
			p.availableThreads <- consumerSignal
		}

		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down job consumer")
			promise <- consumerSignal
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

		p.log.WithFields(logrus.Fields{
			"queue": nextJob.queueName,
			"next":  nextJob.next,
			"now":   p.clock.Now(),
		}).Trace("staged next cron job to be run")

		// TODO What happens if sleep is negative or 0
		// How long do we need to wait for this cron job?
		sleep := nextJob.next.Sub(p.clock.Now())

		// Create a timer.
		timer := time.NewTimer(sleep)
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down cron job consumer")
			promise <- consumerSignal
			timer.Stop()
			return
		case <-timer.C:
			// Bump the cron we just did. But use a slightly more future timestamp.
			// This is to fix a bug where sometimes the cron library seems to be
			// rounding down? Resulting in a `nextTimestamp` that is slightly in the
			// past.
			nextTimestamp := nextJob.schedule.Next(p.clock.Now().Add(900 * time.Millisecond))
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

			// Wrap the execution of the cron
			// This entire block of code is an absolute fucking mess. All it is really
			// doing is wrapping the actual execution of the cron job in a timeout and
			// a sentry span/hub. This way we can attach a ton of useful data to the
			// event when we send it to sentry if it succeeds or fails.
			// TODO Clean it up at some point?
			func(inLog *logrus.Entry, nextJob cronJobTracker) {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					jobTimeoutSeconds*time.Second,
				)
				defer cancel()
				ctx = sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())
				span := sentry.StartSpan(
					ctx,
					"queue.receive",
					sentry.WithTransactionName(nextJob.queueName),
				)
				span.SetData("messaging.message.id", fmt.Sprintf("%s::%s", nextJob.queueName, nextJob.next))
				span.SetData("messaging.destination.name", string(nextJob.queueName))
				span.SetData("messaging.system", "postgresql")
				// For now, sample all cron jobs
				span.Sampled = sentry.SampledTrue

				// Make sure we are logging with the correct context here
				log := inLog.WithContext(span.Context())

				slug := strings.ToLower(nextJob.queueName)
				slug = strings.ReplaceAll(slug, "::", "-")

				hub := sentry.GetHubFromContext(span.Context())
				hub.ConfigureScope(func(scope *sentry.Scope) {
					scope.SetTag("queue", nextJob.queueName)
					scope.SetContext("monitor", sentry.Context{"slug": slug})
				})
				defer hub.RecoverWithContext(span.Context(), nil)

				var err error
				crontab := strings.SplitN(nextJob.handler.DefaultSchedule(), " ", 2)[1]
				monitorSchedule := sentry.CrontabSchedule(crontab)
				monitorConfig := &sentry.MonitorConfig{
					Schedule:      monitorSchedule,
					MaxRuntime:    2,
					CheckInMargin: 1,
				}
				checkInId := hub.CaptureCheckIn(&sentry.CheckIn{
					MonitorSlug: slug,
					Status:      sentry.CheckInStatusInProgress,
				}, monitorConfig)
				defer func() {
					status := sentry.CheckInStatusOK
					span.Status = sentry.SpanStatusOK
					if err != nil {
						status = sentry.CheckInStatusError
						span.Status = sentry.SpanStatusInternalError
						hub.CaptureException(err)
						log.WithError(err).Warn("cron job finished with an error")
					} else {
						log.Debug("cron job finished")
					}
					span.Finish()
					hub.CaptureCheckIn(&sentry.CheckIn{
						ID:          *checkInId,
						MonitorSlug: slug,
						Status:      status,
						Duration:    span.EndTime.Sub(span.StartTime),
					}, monitorConfig)
				}()

				if err = nextJob.handler.EnqueueTriggeredJob(
					span.Context(),
					p.enqueuer,
				); err != nil {
					log.WithError(err).Error("failed to enqueue cron job to be executed")
					return
				}
			}(log, nextJob)

			// If possible, try to trigger the consumption of the cron job locally.
			select {
			case p.trigger <- consumerSignal:
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

func (p *postgresJobProcessor) worker(shutdown chan chan struct{}) {
	for {
		// Tell the consumer that a worker is available.
		p.availableThreads <- consumerSignal
		select {
		case promise := <-shutdown:
			p.log.Debug("shutting down worker thread")
			promise <- consumerSignal
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

			// Execute the job with a timeout.
			ctx, cancel := context.WithTimeout(
				context.Background(),
				jobTimeoutSeconds*time.Second,
			)
			if err := executor(ctx, job); err != nil {
				log.WithError(err).Error("failed to execute job")
			}
			cancel()

			// Try to consume another job again immediately incase there are more.
			p.trigger <- consumerSignal
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
		// We want to have sentry tracking jobs as they are being processed. In
		// order to do this we need to inject a sentry hub into the context and
		// create a new span using that context.
		highContext := sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())
		options := []sentry.SpanOption{
			sentry.WithTransactionName(handler.QueueName()),
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
		span.Description = handler.QueueName()
		span.SetData("input", job.Input)
		span.SetData("messaging.message.id", string(job.JobId))
		span.SetData("messaging.destination.name", string(job.Queue))
		span.SetData("messaging.message.body.size", len(job.Input))
		span.SetData("messaging.message.receive.latency", p.clock.Now().Sub(job.CreatedAt).Milliseconds())
		span.SetData("messaging.system", "postgresql")
		// For now, sample all background jobs
		span.Sampled = sentry.SampledTrue
		jobLog := p.log.WithContext(span.Context()).WithFields(logrus.Fields{
			"jobId": job.JobId,
			"queue": job.Queue,
		})
		hub := sentry.GetHubFromContext(span.Context())
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("queue", handler.QueueName())
		})

		defer func() {
			if panicErr := recover(); panicErr != nil {
				jobLog.Errorf("panic while processing job\n%+v\n%s", panicErr, string(debug.Stack()))
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
				jobLog.WithError(err).Error("error while processing job")
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

			p.markJobStatus(span.Context(), job, err)
		}()
		defer span.Finish()

		jobLog.Trace("handling job")

		// Set err outright to make sentry reporting easier.
		err = handler.HandleConsumeJob(
			span.Context(),
			jobLog,
			[]byte(job.Input),
		)
		return
	}
}
