package mockqueue

import (
	"context"
	"log/slog"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/queue"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/storage"
	"go.uber.org/mock/gomock"
)

var (
	_ queue.Context = &mockContext{}
)

type mockContext struct {
	context *mockgen.MockContext
	db      pg.DBI
}

func NewMockContext(ctx *mockgen.MockContext) queue.Context {
	// Makes it so we don't need stupid shit for tracing
	ctx.EXPECT().Value(gomock.Any()).MinTimes(1)
	ctx.EXPECT().Done().MinTimes(1)
	ctx.EXPECT().Deadline().MinTimes(1)
	return &mockContext{
		context: ctx,
	}
}

// Enqueuer implements [queue.Context].
func (m *mockContext) Enqueuer() queue.Enqueuer {
	return m.context.Enqueuer()
}

// Billing implements [queue.Context].
func (m *mockContext) Billing() billing.Billing {
	return m.context.Billing()
}

// Clock implements [queue.Context].
func (m *mockContext) Clock() clock.Clock {
	return m.context.Clock()
}

// DB implements [queue.Context].
func (m *mockContext) DB() pg.DBI {
	if m.db != nil {
		return m.db
	}
	return m.context.DB()
}

// Deadline implements [queue.Context].
func (m *mockContext) Deadline() (deadline time.Time, ok bool) {
	return m.context.Deadline()
}

// Done implements [queue.Context].
func (m *mockContext) Done() <-chan struct{} {
	return m.context.Done()
}

// Email implements [queue.Context].
func (m *mockContext) Email() communication.EmailCommunication {
	return m.context.Email()
}

// Err implements [queue.Context].
func (m *mockContext) Err() error {
	return m.context.Err()
}

// Job implements [queue.Context].
func (m *mockContext) Job() models.Job {
	return m.context.Job()
}

// KMS implements [queue.Context].
func (m *mockContext) KMS() secrets.KeyManagement {
	return m.context.KMS()
}

// Log implements [queue.Context].
func (m *mockContext) Log() *slog.Logger {
	return m.context.Log()
}

// Platypus implements [queue.Context].
func (m *mockContext) Platypus() platypus.Platypus {
	return m.context.Platypus()
}

// Publisher implements [queue.Context].
func (m *mockContext) Publisher() pubsub.Publisher {
	return m.context.Publisher()
}

// RunInTransaction implements [queue.Context].
func (m *mockContext) RunInTransaction(ctx context.Context, callback func(ctx queue.Context) error) error {
	if err := m.context.RunInTransaction(ctx, callback); err != nil {
		return err
	}

	return m.context.DB().RunInTransaction(ctx, func(tx *pg.Tx) error {
		// TODO this currently doesnt propagate properly
		span := sentry.StartSpan(ctx, "db.transaction")
		defer span.Finish()

		processorClone := *m
		processorClone.db = tx

		return callback(&processorClone)
	})
}

// Storage implements [queue.Context].
func (m *mockContext) Storage() storage.Storage {
	return m.context.Storage()
}

// Value implements [queue.Context].
func (m *mockContext) Value(key any) any {
	return m.context.Value(key)
}
