package background

import (
	"context"
	"testing"
)

type TestJobFunction func(t *testing.T, ctx context.Context, data []byte) error

var (
	_ JobHandler = &TestJobHandler{}
)

type TestJobHandler struct {
	t     *testing.T
	inner TestJobFunction
}

func (h TestJobHandler) QueueName() string {
	return h.t.Name()
}

func (h TestJobHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	return h.inner(h.t, ctx, data)
}

func NewTestJobHandler(t *testing.T, callback TestJobFunction) *TestJobHandler {
	return &TestJobHandler{
		t:     t,
		inner: callback,
	}
}

var (
	_ ScheduledJobHandler = &TestCronJobHandler{}
)

type TestCronJobHandler struct {
	t        *testing.T
	schedule string
	inner    TestJobFunction
}

func (h TestCronJobHandler) DefaultSchedule() string {
	return h.schedule
}

func (h TestCronJobHandler) QueueName() string {
	return h.t.Name()
}

func (h TestCronJobHandler) HandleConsumeJob(ctx context.Context, data []byte) error {
	return h.inner(h.t, ctx, data)
}

func (h TestCronJobHandler) EnqueueTriggeredJob(ctx context.Context, enqueuer JobEnqueuer) error {
	return enqueuer.EnqueueJob(ctx, h.QueueName(), nil)
}

func NewTestCronJobHandler(t *testing.T, schedule string, callback TestJobFunction) *TestCronJobHandler {
	return &TestCronJobHandler{
		t:        t,
		inner:    callback,
		schedule: schedule,
	}
}
