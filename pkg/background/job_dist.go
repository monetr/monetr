package background

import (
	"context"

	"github.com/vmihailenco/msgpack/v5"
)

type JobMarshaller func(v interface{}) ([]byte, error)
type JobUnmarshaller func(src []byte, dst interface{}) error

var (
	DefaultJobMarshaller   JobMarshaller   = msgpack.Marshal
	DefaultJobUnmarshaller JobUnmarshaller = msgpack.Unmarshal
)

type JobHandler interface {
	QueueName() string
	HandleConsumeJob(ctx context.Context, data []byte) error
	SetUnmarshaller(unmarshaller JobUnmarshaller)
}

type ScheduledJobHandler interface {
	JobHandler
	TriggerableJobHandler
	// DefaultSchedule should return the default schedule for this job in cron syntax.
	DefaultSchedule() string
}

type TriggerableJobHandler interface {
	JobHandler
	// EnqueueTriggeredJob should trigger the actual job using the provided enqueuer. Make sure that the triggered job
	// can be run on any worker, not necessarily the worker that triggered the job.
	EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error
}

type Job interface {
	Run(ctx context.Context) error
}

type JobProcessor interface {
	RegisterJob(ctx context.Context, handler JobHandler) error
	Start() error
	Close() error
}

type JobEnqueuer interface {
	EnqueueJob(ctx context.Context, queue string, data interface{}) error
}
