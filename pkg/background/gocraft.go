package background

import (
	"context"
	"encoding/hex"
	"sync/atomic"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ JobEnqueuer  = &GoCraftWorkJobEnqueuer{}
	_ JobProcessor = &GoCraftWorkJobProcessor{}
)

const (
	GoCraftWorkNamespace = "monetr"
)

type GoCraftWorkJobEnqueuer struct {
	log      *logrus.Entry
	enqueuer *work.Enqueuer
	marshal  JobMarshaller
}

func NewGoCraftWorkJobEnqueuer(log *logrus.Entry, redisPool *redis.Pool) *GoCraftWorkJobEnqueuer {
	return &GoCraftWorkJobEnqueuer{
		log:      log,
		enqueuer: work.NewEnqueuer(GoCraftWorkNamespace, redisPool),
		marshal:  DefaultJobMarshaller,
	}
}

func (g *GoCraftWorkJobEnqueuer) EnqueueJob(ctx context.Context, queue string, arguments interface{}) error {
	span := sentry.StartSpan(ctx, "topic.send")
	defer span.Finish()

	span.Description = "gocraft Enqueue"
	span.SetTag("queue", queue)
	span.Data = map[string]interface{}{
		"queue":     queue,
		"arguments": arguments,
	}

	crumbs.Debug(span.Context(), "Enqueueing job for background processing", map[string]interface{}{
		"queue":     queue,
		"arguments": arguments,
	})

	g.log.WithContext(span.Context()).WithField("queue", queue).Debug("enqueuing job to be run")

	encodedArguments, err := g.marshal(arguments)
	if err = errors.Wrap(err, "failed to marshal arguments"); err != nil {
		return err
	}

	_, err = g.enqueuer.EnqueueUnique(queue, map[string]interface{}{
		"args": hex.EncodeToString(encodedArguments),
	})
	if err = errors.Wrap(err, "failed to enqueue job for gocraft/work"); err != nil {
		return err
	}

	return nil
}

type GoCraftWorkJobProcessor struct {
	state         uint32
	configuration config.BackgroundJobs
	log           *logrus.Entry
	queue         *work.WorkerPool
	enqueuer      JobEnqueuer
	marshal       JobMarshaller
}

func NewGoCraftWorkJobProcessor(
	log *logrus.Entry,
	configuration config.BackgroundJobs,
	redisPool *redis.Pool,
	enqueuer JobEnqueuer,
) *GoCraftWorkJobProcessor {
	return &GoCraftWorkJobProcessor{
		configuration: configuration,
		log:           log,
		queue:         work.NewWorkerPool(struct{}{}, 4, GoCraftWorkNamespace, redisPool),
		enqueuer:      enqueuer,
		marshal:       DefaultJobMarshaller,
	}
}

func (g *GoCraftWorkJobProcessor) Start() error {
	if !atomic.CompareAndSwapUint32(&g.state, 0, 1) {
		return errors.New("gocraft/work job processor is either already started, or is in an invalid state")
	}

	g.queue.Start()
	return nil
}

func (g *GoCraftWorkJobProcessor) Close() error {
	if !atomic.CompareAndSwapUint32(&g.state, 1, 2) {
		return errors.New("gocraft/work job processor is either already closed, or is in an invalid state")
	}

	doneChannel := make(chan struct{})
	timeout := time.NewTimer(30 * time.Second)

	// queue.Drain can technically run forever. So run it in a go-routine and defer the done signal for when it finishes
	// if it ever finishes. If it does not then it will be killed if we timeout anyway.
	go func() {
		defer func() {
			doneChannel <- struct{}{}
		}()
		g.queue.Drain()
	}()

	defer func() {
		g.queue.Stop()
	}()
	select {
	case <-doneChannel:
		return nil
	case <-timeout.C:
		return errors.New("timeout while draining gocraft/work queue")
	}
}

func (g *GoCraftWorkJobProcessor) RegisterJob(ctx context.Context, handler JobHandler) error {
	if atomic.LoadUint32(&g.state) == 2 {
		return errors.New("gocraft/work job processor is either closed or is in an invalid state")
	}

	log := g.log.WithContext(ctx).WithField("job", handler.QueueName())
	log.Trace("registering job handler")
	g.queue.Job(handler.QueueName(), func(job *work.Job) (err error) {
		// We want to have sentry tracking jobs as they are being processed. In order to do this we need to inject a
		// sentry hub into the context and create a new span using that context.
		highContext := sentry.SetHubOnContext(context.Background(), sentry.CurrentHub().Clone())
		span := sentry.StartSpan(highContext, "queue.process", sentry.TransactionName(handler.QueueName()))
		jobLog := log.WithContext(span.Context())
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
		}()
		defer span.Finish()

		jobLog.Trace("handling job")

		var data []byte
		if argsString, ok := job.Args["args"].(string); ok {
			data, err = hex.DecodeString(argsString)
			if err = errors.Wrap(err, "failed to decode args string"); err != nil {
				jobLog.WithError(err).Error("failed to decode arguments for job")
				return err
			}
		}

		err = handler.HandleConsumeJob(span.Context(), data) // Set err outright to make sentry reporting easier.
		return
	})

	// When we are using the internal scheduler (AKA the built-in gocraft scheduler) we need to set up each job that has
	// a schedule with the gocraft scheduler.
	if g.configuration.Scheduler == config.BackgroundJobSchedulerInternal {
		if scheduledJob, ok := handler.(ScheduledJobHandler); ok {
			schedule := scheduledJob.DefaultSchedule()
			log.WithField("schedule", schedule).Trace("job will be run on a schedule automatically")
			// We use a special queue name to trigger the actual jobs.
			schedulerName := scheduledJob.QueueName() + "GoCraftScheduler"
			// Each job is wrapped with an enqueuer. This enqueuer is run on a schedule to trigger the actual jobs in
			// their batches.
			g.queue.Job(schedulerName, func(job *work.Job) error {
				return scheduledJob.EnqueueTriggeredJob(context.Background(), g.enqueuer)
			})
			g.queue.PeriodicallyEnqueue(scheduledJob.DefaultSchedule(), schedulerName)
		}
	}

	return nil
}
