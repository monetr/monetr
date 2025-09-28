package background

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type JobMarshaller func(v any) ([]byte, error)
type JobUnmarshaller func(src []byte, dst any) error

var (
	DefaultJobMarshaller   JobMarshaller   = json.Marshal
	DefaultJobUnmarshaller JobUnmarshaller = json.Unmarshal
)

type JobHandler interface {
	QueueName() string
	HandleConsumeJob(ctx context.Context, log *logrus.Entry, data []byte) error
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

type JobImplementation interface {
	Run(ctx context.Context) error
}

type JobProcessor interface {
	RegisterJob(ctx context.Context, handler JobHandler) error
	Start() error
	Close() error
}
