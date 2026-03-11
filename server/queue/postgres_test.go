package queue

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testJobArgs struct {
	Value string `json:"value"`
}

func testNoopJob(ctx Context, args testJobArgs) error {
	return nil
}

func testFailingJob(ctx Context, args testJobArgs) error {
	return errors.New("this job always fails")
}

func testNoopCron(ctx Context) error {
	return nil
}

func newTestProcessor(t *testing.T, clock clock.Clock, notifier Notifier, databaseOptions ...testutils.DatabaseOption) Processor {
	t.Helper()
	db := testutils.GetPgDatabase(t, databaseOptions...)
	log := testutils.GetLog(t)
	return NewPostgresQueue(
		t.Context(),
		notifier,
		clock,
		log,
		config.Configuration{},
		db,
		nil, // publisher
		nil, // platypus
		nil, // kms
		nil, // storage
		nil, // billing
		nil, // email
	)
}

func TestPostgresProcessor_Register(t *testing.T) {
	t.Run("can register a job", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, NewMemoryNotifier(4))

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err, "must be able to register a job handler")
	})

	t.Run("cannot register the same job twice", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, NewMemoryNotifier(4))

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err, "first registration must succeed")

		err = Register(t.Context(), processor, testNoopJob)
		assert.Error(t, err, "second registration of the same job must return an error")
	})

	t.Run("can register a cron job", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, NewMemoryNotifier(4))

		err := RegisterCron(t.Context(), processor, "0 0 * * * *", testNoopCron)
		assert.NoError(t, err, "must be able to register a cron job handler")
	})

	t.Run("cannot register the same cron job twice", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, NewMemoryNotifier(4))

		err := RegisterCron(t.Context(), processor, "0 0 * * * *", testNoopCron)
		assert.NoError(t, err, "first registration must succeed")

		err = RegisterCron(t.Context(), processor, "0 0 * * * *", testNoopCron)
		assert.Error(t, err, "second registration of the same cron must return an error")
	})

	t.Run("cannot register a job after the processor has started", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, NewMemoryNotifier(4), testutils.IsolatedDatabase)

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err)

		err = processor.Start()
		assert.NoError(t, err, "must be able to start the processor")
		defer processor.Close()

		err = Register(t.Context(), processor, testFailingJob)
		assert.Error(t, err, "must not be able to register a job after the processor has started")
	})
}

func TestPostgresProcessor_Start(t *testing.T) {
	t.Run("cannot start with no jobs registered", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, NewMemoryNotifier(4))

		err := processor.Start()
		assert.Error(t, err, "must not be able to start a processor with no jobs registered")
	})

	t.Run("cannot start a processor that is already running", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, NewMemoryNotifier(4), testutils.IsolatedDatabase)

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err)

		err = processor.Start()
		assert.NoError(t, err, "must be able to start the processor")
		defer processor.Close()

		err = processor.Start()
		assert.Error(t, err, "starting an already-running processor must return an error")
	})

	t.Run("cannot close a processor that has not been started", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, NewMemoryNotifier(4))

		err := processor.Close()
		assert.Error(t, err, "closing a processor that has not been started must return an error")
	})
}

func TestPostgresProcessor_ExecuteJob(t *testing.T) {
	t.Run("successful job is marked as completed", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			NewMemoryNotifier(4),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.register(t.Context(), "test-noop", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		now := clock.Now()
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test-noop",
			Signature: "test-sig-noop",
			Priority:  uint64(now.Unix()),
			Input:     `{}`,
			Status:    models.ProcessingJobStatus,
			Attempt:   1,
			CreatedAt: now,
			UpdatedAt: now,
		})

		p.executeJob(&job)

		updated := testutils.MustDBRead(t, job)
		assert.Equal(t, models.CompletedJobStatus, updated.Status, "job must be marked as completed")
		assert.NotNil(t, updated.CompletedAt, "completed_at must be set")
	})

	t.Run("failing job is rescheduled for retry", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			NewMemoryNotifier(4),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.register(t.Context(), "test-failing", func(_ Context, _ []byte) error {
			return errors.New("always fails")
		})
		assert.NoError(t, err)

		now := clock.Now()
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test-failing",
			Signature: "test-sig-failing",
			Priority:  uint64(now.Unix()),
			Input:     `{}`,
			Status:    models.ProcessingJobStatus,
			Attempt:   1,
			CreatedAt: now,
			UpdatedAt: now,
		})

		p.executeJob(&job)

		updated := testutils.MustDBRead(t, job)
		assert.Equal(t, models.PendingJobStatus, updated.Status, "job must be rescheduled as pending for retry")
		assert.EqualValues(t, 2, updated.Attempt, "attempt count must be incremented")
		assert.Nil(t, updated.StartedAt, "started_at must be cleared for retry")
	})

	t.Run("job with no remaining attempts is marked as failed", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			NewMemoryNotifier(4),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.register(t.Context(), "test-exhausted", func(_ Context, _ []byte) error {
			return errors.New("always fails")
		})
		assert.NoError(t, err)

		now := clock.Now()
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test-exhausted",
			Signature: "test-sig-exhausted",
			Priority:  uint64(now.Unix()),
			Input:     `{}`,
			Status:    models.ProcessingJobStatus,
			Attempt:   maxAttempts, // no attempts remaining
			CreatedAt: now,
			UpdatedAt: now,
		})

		p.executeJob(&job)

		updated := testutils.MustDBRead(t, job)
		assert.Equal(t, models.FailedJobStatus, updated.Status, "job must be marked as failed after exhausting all attempts")
		assert.NotNil(t, updated.CompletedAt, "completed_at must be set when a job is failed")
	})
}

func TestPostgresProcessor_EnqueueAndExecute(t *testing.T) {
	t.Run("job is executed and marked as completed", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		processor := NewPostgresQueue(
			t.Context(),
			NewMemoryNotifier(4),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		)

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err)

		err = processor.Start()
		assert.NoError(t, err, "must be able to start the processor")
		defer processor.Close()

		err = Enqueue(t.Context(), processor, testNoopJob, testJobArgs{Value: "hello"})
		assert.NoError(t, err, "must be able to enqueue a job")

		require.Eventually(t, func() bool {
			count, err := db.Model(new(models.Job)).
				Where(`"status" = ?`, models.CompletedJobStatus).
				Count()
			return err == nil && count > 0
		}, 10*time.Second, 100*time.Millisecond, "job must be executed and marked as completed")
	})

	t.Run("duplicate enqueue within the same second is silently dropped", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		processor := NewPostgresQueue(
			t.Context(),
			NewMemoryNotifier(4),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		)

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err)

		// Use a fixed timestamp so both enqueues have the same signature.
		at := time.Now().Truncate(time.Second)
		args := testJobArgs{Value: "deduplicated"}

		err = EnqueueAt(t.Context(), processor, at, testNoopJob, args)
		assert.NoError(t, err, "first enqueue must succeed")

		err = EnqueueAt(t.Context(), processor, at, testNoopJob, args)
		assert.NoError(t, err, "duplicate enqueue must not return an error")

		count, err := db.Model(new(models.Job)).Count()
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "only one job row must exist despite two enqueue calls")
	})

	t.Run("cron job fires and is executed", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		processor := NewPostgresQueue(
			t.Context(),
			NewMemoryNotifier(4),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		)

		// Every second so the test doesn't wait long.
		err := RegisterCron(t.Context(), processor, "* * * * * *", testNoopCron)
		assert.NoError(t, err)

		err = processor.Start()
		assert.NoError(t, err, "must be able to start the processor")
		defer processor.Close()

		require.Eventually(t, func() bool {
			count, err := db.Model(new(models.Job)).
				Where(`"status" = ?`, models.CompletedJobStatus).
				Count()
			return err == nil && count > 0
		}, 5*time.Second, 100*time.Millisecond, "cron job must fire and be marked as completed")
	})
}
