package background

import (
	"context"
	"testing"
)

type TestJobFunction func(t *testing.T, ctx context.Context, data []byte) error

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
