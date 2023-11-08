package background

import (
	"context"
	"encoding/hex"
	"sync/atomic"
	"time"

	"github.com/acaloiaro/neoq"
	"github.com/acaloiaro/neoq/backends/postgres"
	"github.com/acaloiaro/neoq/handler"
	"github.com/acaloiaro/neoq/jobs"
	neoqlog "github.com/acaloiaro/neoq/logging"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/logging"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ JobEnqueuer = &NeoqJobEnqueuer{}
)

type NeoqJobEnqueuer struct {
	log     *logrus.Entry
	nq      neoq.Neoq
	marshal JobMarshaller
}

func NewNeoqJobEnqueuerPostgres(log *logrus.Entry, connectionString string) (*NeoqJobEnqueuer, error) {
	nq, err := neoq.New(
		context.Background(),
		neoq.WithBackend(postgres.Backend),
		postgres.WithConnectionString(connectionString),
		neoq.WithLogLevel(neoqlog.LogLevelDebug),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize neoq with postgres backend")
	}
	nq.SetLogger(logging.NewLogrusWrapper(log))
	return &NeoqJobEnqueuer{
		log:     log,
		nq:      nq,
		marshal: DefaultJobMarshaller,
	}, nil
}

func (n *NeoqJobEnqueuer) EnqueueJob(ctx context.Context, queue string, arguments interface{}) error {
	span := sentry.StartSpan(ctx, "topic.send")
	defer span.Finish()
	span.Description = "neoq Enqueue"
	span.SetTag("queue", queue)
	span.Data = map[string]interface{}{
		"queue":     queue,
		"arguments": arguments,
	}

	crumbs.Debug(span.Context(), "Enqueueing job for background processing", map[string]interface{}{
		"queue":     queue,
		"arguments": arguments,
	})

	n.log.WithContext(span.Context()).WithField("queue", queue).Debug("enqueuing job to be run")

	encodedArguments, err := n.marshal(arguments)
	if err = errors.Wrap(err, "failed to marshal arguments"); err != nil {
		return err
	}

	jobId, err := n.nq.Enqueue(span.Context(), &jobs.Job{
		Queue: queue,
		Payload: map[string]any{
			"args": hex.EncodeToString(encodedArguments),
		},
	})
	if err = errors.Wrap(err, "failed to enqueue job for neoq"); err != nil {
		return err
	}
	span.Data["jobId"] = jobId

	return nil
}

type NeoqJobProcessor struct {
	state         uint32
	configuration config.BackgroundJobs
	log           *logrus.Entry
	nq            neoq.Neoq
	enqueuer      JobEnqueuer
	marshal       JobMarshaller
}

func NewNeoqJobProcessorPostgres(
	log *logrus.Entry,
	configuration config.BackgroundJobs,
	connectionString string,
	enqueuer JobEnqueuer,
) (*NeoqJobProcessor, error) {
	nq, err := neoq.New(
		context.Background(),
		neoq.WithBackend(postgres.Backend),
		postgres.WithConnectionString(connectionString),
		postgres.WithSynchronousCommit(false),
		neoq.WithLogLevel(neoqlog.LogLevelDebug),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize neoq with postgres backend")
	}
	nq.SetLogger(logging.NewLogrusWrapper(log))
	return &NeoqJobProcessor{
		configuration: configuration,
		log:           log,
		nq:            nq,
		enqueuer:      enqueuer,
		marshal:       DefaultJobMarshaller,
	}, nil
}

func (n *NeoqJobProcessor) Start() error {
	if !atomic.CompareAndSwapUint32(&n.state, 0, 1) {
		return errors.New("neoq job processor is either already started, or is in an invalid state")
	}

	// No-op for this processor, it automatically starts on new.
	return nil
}

func (n *NeoqJobProcessor) Close() error {
	if !atomic.CompareAndSwapUint32(&n.state, 1, 2) {
		return errors.New("gocraft/work job processor is either already closed, or is in an invalid state")
	}

	doneChannel := make(chan struct{})
	timeout := time.NewTimer(30 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// queue.Drain can technically run forever. So run it in a go-routine and defer the done signal for when it finishes
	// if it ever finishes. If it does not then it will be killed if we timeout anyway.
	go func() {
		defer func() {
			doneChannel <- struct{}{}
		}()
		n.nq.Shutdown(ctx)
	}()

	select {
	case <-doneChannel:
		return nil
	case <-timeout.C:
		return errors.New("timeout while shutting down the neoq queue")
	}
}

func (n *NeoqJobProcessor) RegisterJob(ctx context.Context, jobHandler JobHandler) error {
	if atomic.LoadUint32(&n.state) == 2 {
		return errors.New("neoq job processor is either closed or is in an invalid state")
	}

	log := n.log.WithContext(ctx).WithField("job", jobHandler.QueueName())
	if n.configuration.Scheduler == config.BackgroundJobSchedulerInternal {
		if scheduledJob, ok := jobHandler.(ScheduledJobHandler); ok {
			schedule := scheduledJob.DefaultSchedule()
			log.WithField("schedule", schedule).Trace("job will be run on a schedule automatically")
			// We use a special queue name to trigger the actual jobs.
			schedulerName := scheduledJob.QueueName() + "NeoqScheduler"
			// Each job is wrapped with an enqueuer. This enqueuer is run on a schedule to trigger the actual jobs in
			// their batches.

			cronHandler := handler.NewPeriodic(
				func(ctx context.Context) error {
					span := sentry.StartSpan(
						ctx,
						"topic.process",
						sentry.TransactionName(schedulerName),
					)
					defer span.Finish()
					return scheduledJob.EnqueueTriggeredJob(span.Context(), n.enqueuer)
				},
				handler.Queue(schedulerName),
				handler.Concurrency(1),
			)

			if err := n.nq.StartCron(context.Background(), schedule, cronHandler); err != nil {
				return errors.Wrap(err, "failed to register cron job")
			}
			time.Sleep(50 * time.Millisecond)
			log.WithField("schedule", schedule).Trace("sucessfully registered job to run on a schedule")
		}
	}

	log.Trace("registering non-scheduled job handler")
	neoqJobHandler := createNeoqJobHandler(log, jobHandler)
	err := n.nq.Start(ctx, neoqJobHandler)
	if err != nil {
		return errors.Wrap(err, "failed to register job with neoq")
	}
	log.Trace("successfully registered non-scheduled job handler")

	return nil
}

func createNeoqJobHandler(log *logrus.Entry, jobHandler JobHandler) handler.Handler {
	neoqHandler := handler.New(jobHandler.QueueName(), func(ctx context.Context) (err error) {
		// We want to have sentry tracking jobs as they are being processed. In order to do this we need to inject a
		// sentry hub into the context and create a new span using that context.
		highContext := sentry.SetHubOnContext(ctx, sentry.CurrentHub().Clone())
		span := sentry.StartSpan(highContext, "topic.process", sentry.TransactionName(jobHandler.QueueName()))
		span.Description = jobHandler.QueueName()
		jobLog := log.WithContext(span.Context())
		hub := sentry.GetHubFromContext(span.Context())
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("queue", jobHandler.QueueName())
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
		job, err := jobs.FromContext(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to extract job details from context")
		}
		if argsString, ok := job.Payload["args"].(string); ok {
			data, err = hex.DecodeString(argsString)
			if err = errors.Wrap(err, "failed to decode args string"); err != nil {
				jobLog.WithError(err).Error("failed to decode arguments for job")
				return err
			}
		}

		err = jobHandler.HandleConsumeJob(span.Context(), data) // Set err outright to make sentry reporting easier.
		return
	})

	return neoqHandler
}
