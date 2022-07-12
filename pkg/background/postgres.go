package background

import (
	"context"
	"encoding/json"
	"sync/atomic"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/postgresque"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	_ JobEnqueuer = &PostgresJobEnqueuer{}
	_ JobProcessor = &PostgresJobProcessor{}
)

type PostgresJobEnqueuer struct {
	log     *logrus.Entry
	queue   *postgresque.Queue
	marshal JobMarshaller
}

func NewPostgresJobEnqueuer(log *logrus.Entry, queue *postgresque.Queue) *PostgresJobEnqueuer {
	return &PostgresJobEnqueuer{
		log:     log,
		queue:   queue,
		marshal: json.Marshal, // Not used here.
	}
}

func (p *PostgresJobEnqueuer) EnqueueJob(ctx context.Context, queue string, arguments interface{}) error {
	span := sentry.StartSpan(ctx, "topic.send")
	defer span.Finish()

	span.Description = "postgres Enqueue"
	span.SetTag("queue", queue)
	span.Data = map[string]interface{}{
		"queue":     queue,
		"arguments": arguments,
	}

	crumbs.Debug(span.Context(), "Enqueueing job for background processing", map[string]interface{}{
		"queue":     queue,
		"arguments": arguments,
	})

	log := p.log.WithContext(span.Context()).WithField("queue", queue)

	log.Debug("enqueuing job to be run")

	jobId, err := p.queue.Enqueue(span.Context(), queue, arguments)
	if err != nil {
		return errors.Wrap(err, "failed to enqueue job")
	}

	log.WithField("jobId", jobId).Debug("successfully enqueued job")

	return nil
}

type PostgresJobProcessor struct {
	state         uint32
	configuration config.BackgroundJobs
	log           *logrus.Entry
	queue         *postgresque.Queue
	enqueuer      JobEnqueuer
	marshal       JobMarshaller
}

func NewPostgresJobProcessor(
	log *logrus.Entry,
	configuration config.BackgroundJobs,
	queue *postgresque.Queue,
	enqueuer JobEnqueuer,
) *PostgresJobProcessor {
	return &PostgresJobProcessor{
		configuration: configuration,
		log:           log,
		queue:         queue,
		enqueuer:      enqueuer,
		marshal:       json.Marshal,
	}
}

func (p *PostgresJobProcessor) RegisterJob(ctx context.Context, handler JobHandler) error {
	if atomic.LoadUint32(&p.state) == 2 {
		return errors.New("postgres job processor is either closed or is in an invalid state")
	}

	log := p.log.WithContext(ctx).WithField("job", handler.QueueName())
	log.Trace("registering job handler")

	panic("not implemented")
}

func (p *PostgresJobProcessor) Start() error {
	//TODO implement me
	panic("implement me")
}

func (p *PostgresJobProcessor) Close() error {
	//TODO implement me
	panic("implement me")
}
