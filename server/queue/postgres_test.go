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

func newTestProcessor(t *testing.T, clock clock.Clock, databaseOptions ...testutils.DatabaseOption) Processor {
	t.Helper()
	db := testutils.GetPgDatabase(t, databaseOptions...)
	log := testutils.GetLog(t)
	return NewPostgresQueue(
		t.Context(),
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
		processor := newTestProcessor(t, clock)

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err, "must be able to register a job handler")
	})

	t.Run("cannot register the same job twice", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock)

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err, "first registration must succeed")

		err = Register(t.Context(), processor, testNoopJob)
		assert.Error(t, err, "second registration of the same job must return an error")
	})

	t.Run("can register a cron job", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock)

		err := RegisterCron(t.Context(), processor, testNoopCron, "0 0 * * * *")
		assert.NoError(t, err, "must be able to register a cron job handler")
	})

	t.Run("cannot register the same cron job twice", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock)

		err := RegisterCron(t.Context(), processor, testNoopCron, "0 0 * * * *")
		assert.NoError(t, err, "first registration must succeed")

		err = RegisterCron(t.Context(), processor, testNoopCron, "0 0 * * * *")
		assert.Error(t, err, "second registration of the same cron must return an error")
	})

	t.Run("cannot register a job after the processor has started", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, testutils.IsolatedDatabase)

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
		processor := newTestProcessor(t, clock)

		err := processor.Start()
		assert.Error(t, err, "must not be able to start a processor with no jobs registered")
	})

	t.Run("cannot start a processor that is already running", func(t *testing.T) {
		clock := clock.NewMock()
		processor := newTestProcessor(t, clock, testutils.IsolatedDatabase)

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
		processor := newTestProcessor(t, clock)

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
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-noop", func(_ Context, _ []byte) error {
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
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-failing", func(_ Context, _ []byte) error {
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
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-exhausted", func(_ Context, _ []byte) error {
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

	t.Run("panicking job is marked as failed without retry", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-panic", func(_ Context, _ []byte) error {
			panic("something went terribly wrong")
		})
		assert.NoError(t, err)

		now := clock.Now()
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test-panic",
			Signature: "test-sig-panic",
			Priority:  uint64(now.Unix()),
			Input:     `{}`,
			Status:    models.ProcessingJobStatus,
			Attempt:   1,
			CreatedAt: now,
			UpdatedAt: now,
		})

		p.executeJob(&job)

		updated := testutils.MustDBRead(t, job)
		assert.Equal(t, models.FailedJobStatus, updated.Status, "panicking job must be marked as failed")
		assert.NotNil(t, updated.CompletedAt, "completed_at must be set")
		assert.EqualValues(t, 1, updated.Attempt, "attempt must not be incremented for panics")
	})

	t.Run("retry backoff increases priority correctly", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-backoff", func(_ Context, _ []byte) error {
			return errors.New("always fails")
		})
		assert.NoError(t, err)

		now := clock.Now()
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test-backoff",
			Signature: "test-sig-backoff",
			Priority:  uint64(now.Unix()),
			Input:     `{}`,
			Status:    models.ProcessingJobStatus,
			Attempt:   2,
			CreatedAt: now,
			UpdatedAt: now,
		})

		p.executeJob(&job)

		updated := testutils.MustDBRead(t, job)
		expectedPriority := uint64(now.Add(attemptBackoff * 2).Unix())
		assert.Equal(t, models.PendingJobStatus, updated.Status, "job must be rescheduled as pending")
		assert.Equal(t, expectedPriority, updated.Priority, "priority must be bumped by attemptBackoff * attempt")
	})

	t.Run("successive retries increment attempt each time", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-successive", func(_ Context, _ []byte) error {
			return errors.New("always fails")
		})
		assert.NoError(t, err)

		now := clock.Now()
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test-successive",
			Signature: "test-sig-successive",
			Priority:  uint64(now.Unix()),
			Input:     `{}`,
			Status:    models.ProcessingJobStatus,
			Attempt:   1,
			CreatedAt: now,
			UpdatedAt: now,
		})

		p.executeJob(&job)

		updated := testutils.MustDBRead(t, job)
		assert.Equal(t, models.PendingJobStatus, updated.Status)
		assert.EqualValues(t, 2, updated.Attempt)

		// Simulate the job being picked up again for a second attempt.
		updated.Status = models.ProcessingJobStatus
		testutils.MustDBUpdate(t, &updated)

		p.executeJob(&updated)

		updated = testutils.MustDBRead(t, job)
		assert.Equal(t, models.PendingJobStatus, updated.Status)
		assert.EqualValues(t, 3, updated.Attempt)
	})
}

func TestPostgresProcessor_EnqueueAndExecute(t *testing.T) {
	t.Run("job is executed and marked as completed", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		processor := NewPostgresQueue(
			t.Context(),
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
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		)

		err := Register(t.Context(), processor, testNoopJob)
		assert.NoError(t, err)

		// Use a fixed timestamp so both enqueues have the same signature.
		at := clock.Now().Truncate(time.Second)
		args := testJobArgs{Value: "deduplicated"}

		err = EnqueueAt(t.Context(), processor, at, testNoopJob, args)
		assert.NoError(t, err, "first enqueue must succeed")

		err = EnqueueAt(t.Context(), processor, at, testNoopJob, args)
		assert.NoError(t, err, "duplicate enqueue must not return an error")

		count, err := db.Model(new(models.Job)).Count()
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "only one job row must exist despite two enqueue calls")
	})

	t.Run("different queues with same timestamp and args are not deduplicated", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "cron-a", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		err = p.Register(t.Context(), "cron-b", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		// Use the same timestamp and nil args, exactly like the cron consumer
		// does when two cron jobs share the same schedule.
		at := clock.Now().Truncate(time.Second)
		err = p.EnqueueAt(t.Context(), "cron-a", at, nil)
		assert.NoError(t, err, "first queue enqueue must succeed")

		err = p.EnqueueAt(t.Context(), "cron-b", at, nil)
		assert.NoError(t, err, "second queue enqueue must succeed")

		count, err := db.Model(new(models.Job)).Count()
		assert.NoError(t, err)
		assert.Equal(t, 2, count, "both jobs must exist because they belong to different queues")
	})

	t.Run("cron job fires and is executed", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		processor := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		)

		// Every second so the test doesn't wait long.
		err := RegisterCron(t.Context(), processor, testNoopCron, "* * * * * *")
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

func TestPostgresProcessor_EnqueueAt(t *testing.T) {
	t.Run("enqueue sets initial attempt to 1", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-enqueue", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		at := clock.Now().Truncate(time.Second)
		err = p.EnqueueAt(t.Context(), "test-enqueue", at, testJobArgs{Value: "test"})
		assert.NoError(t, err, "must be able to enqueue a job")

		var job models.Job
		err = db.Model(&job).Where(`"queue" = ?`, "test-enqueue").Select()
		require.NoError(t, err)
		assert.EqualValues(t, 1, job.Attempt, "initial attempt must be 1")
		assert.Equal(t, models.PendingJobStatus, job.Status, "initial status must be pending")
	})

	t.Run("enqueue sets priority to the unix timestamp of the at parameter", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-priority", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		at := time.Date(2026, 6, 15, 12, 30, 0, 0, time.UTC)
		err = p.EnqueueAt(t.Context(), "test-priority", at, testJobArgs{Value: "test"})
		assert.NoError(t, err, "must be able to enqueue a job")

		var job models.Job
		err = db.Model(&job).Where(`"queue" = ?`, "test-priority").Select()
		require.NoError(t, err)
		assert.Equal(t, uint64(at.Unix()), job.Priority, "priority must match the at timestamp")
	})
}

func TestPostgresProcessor_ConsumeJobMaybe(t *testing.T) {
	t.Run("consumes pending job with past priority", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-consume", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		now := clock.Now()
		testutils.MustInsert(t, models.Job{
			Queue:     "test-consume",
			Signature: "test-sig-consume",
			Priority:  uint64(now.Add(-1 * time.Hour).Unix()),
			Input:     `{}`,
			Status:    models.PendingJobStatus,
			Attempt:   1,
			CreatedAt: now,
			UpdatedAt: now,
		})

		job, err := p.consumeJobMaybe()
		assert.NoError(t, err)
		require.NotNil(t, job, "must consume a pending job with past priority")
		assert.Equal(t, models.ProcessingJobStatus, job.Status, "consumed job must be in processing status")
		assert.NotNil(t, job.StartedAt, "started_at must be set on consumed job")
	})

	t.Run("does not consume a job already in processing status", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-no-consume", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		now := clock.Now()
		startedAt := now.Add(-5 * time.Minute)
		testutils.MustInsert(t, models.Job{
			Queue:     "test-no-consume",
			Signature: "test-sig-no-consume",
			Priority:  uint64(now.Add(-1 * time.Hour).Unix()),
			Input:     `{}`,
			Status:    models.ProcessingJobStatus,
			Attempt:   1,
			CreatedAt: now,
			UpdatedAt: now,
			StartedAt: &startedAt,
		})

		job, err := p.consumeJobMaybe()
		assert.NoError(t, err)
		assert.Nil(t, job, "must not consume a job that is already processing")
	})

	t.Run("does not consume a job for an unregistered queue", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-registered", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		now := clock.Now()
		testutils.MustInsert(t, models.Job{
			Queue:     "test-unregistered",
			Signature: "test-sig-unregistered",
			Priority:  uint64(clock.Now().Add(-1 * time.Hour).Unix()),
			Input:     `{}`,
			Status:    models.PendingJobStatus,
			Attempt:   1,
			CreatedAt: now,
			UpdatedAt: now,
		})

		job, err := p.consumeJobMaybe()
		assert.NoError(t, err)
		assert.Nil(t, job, "must not consume a job for an unregistered queue")
	})

	t.Run("returns nil when no jobs exist", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-empty", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		job, err := p.consumeJobMaybe()
		assert.NoError(t, err)
		assert.Nil(t, job, "must return nil when no jobs exist")
	})

	t.Run("does not consume a job with future priority", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-future", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		testutils.MustInsert(t, models.Job{
			Queue:     "test-future",
			Signature: "test-sig-future",
			Priority:  uint64(clock.Now().Add(1 * time.Hour).Unix()),
			Input:     `{}`,
			Status:    models.PendingJobStatus,
			Attempt:   1,
			CreatedAt: clock.Now(),
			UpdatedAt: clock.Now(),
		})

		job, err := p.consumeJobMaybe()
		assert.NoError(t, err)
		assert.Nil(t, job, "must not consume a job with future priority")
	})
}

func TestPostgresProcessor_ConsumeCronMaybe(t *testing.T) {
	t.Run("consumes a cron job whose next_run_at is in the past", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		testutils.MustInsert(t, models.CronJob{
			Queue:        "test-cron-consume",
			CronSchedule: "0 0 * * * *",
			NextRunAt:    clock.Now().Add(-1 * time.Hour),
		})

		nextRun := clock.Now().Add(1 * time.Hour)
		cronJob, err := p.consumeCronMaybe("test-cron-consume", nextRun)
		assert.NoError(t, err)
		require.NotNil(t, cronJob, "must consume a cron job with past next_run_at")

		updated := testutils.MustDBRead(t, models.CronJob{Queue: "test-cron-consume"})
		assert.NotNil(t, updated.LastRunAt, "last_run_at must be set after consumption")
	})

	t.Run("does not consume a cron job whose next_run_at is in the future", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		now := clock.Now()
		testutils.MustInsert(t, models.CronJob{
			Queue:        "test-cron-future",
			CronSchedule: "0 0 * * * *",
			NextRunAt:    now.Add(1 * time.Hour),
		})

		// Pass a next time that is before the cron's next_run_at.
		cronJob, err := p.consumeCronMaybe("test-cron-future", now.Add(30*time.Minute))
		assert.NoError(t, err)
		assert.Nil(t, cronJob, "must not consume a cron job whose next_run_at is in the future")
	})
}

func TestPostgresProcessor_HydrateCronJobTable(t *testing.T) {
	t.Run("removes cron jobs that are no longer registered", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.RegisterCron(t.Context(), "test-current", "0 0 * * * *", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		// Insert a stale cron row that is not registered on this processor.
		testutils.MustInsert(t, models.CronJob{
			Queue:        "test-stale",
			CronSchedule: "0 0 * * * *",
			NextRunAt:    clock.Now(),
		})

		err = p.hydrateCronJobTable()
		assert.NoError(t, err, "hydrate must not return an error")

		// The stale cron must have been removed.
		staleExists, err := db.Model(new(models.CronJob)).
			Where(`"queue" = ?`, "test-stale").
			Exists()
		assert.NoError(t, err)
		assert.False(t, staleExists, "stale cron job must be removed")

		// The registered cron must exist.
		currentExists, err := db.Model(new(models.CronJob)).
			Where(`"queue" = ?`, "test-current").
			Exists()
		assert.NoError(t, err)
		assert.True(t, currentExists, "registered cron job must exist")
	})

	t.Run("inserts new cron jobs for newly registered queues", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.RegisterCron(t.Context(), "test-new-cron", "0 0 * * * *", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		err = p.hydrateCronJobTable()
		assert.NoError(t, err, "hydrate must not return an error")

		var cronJob models.CronJob
		err = db.Model(&cronJob).Where(`"queue" = ?`, "test-new-cron").Select()
		require.NoError(t, err, "cron job row must exist after hydration")
		assert.Equal(t, "0 0 * * * *", cronJob.CronSchedule, "schedule must match the registered schedule")
		assert.True(t, cronJob.NextRunAt.After(clock.Now()), "next_run_at must be in the future relative to the clock")
	})

	t.Run("updates schedule when cron expression changes", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		// Register with the new schedule.
		err := p.RegisterCron(t.Context(), "test-update-cron", "0 0 * * * *", func(_ Context, _ []byte) error {
			return nil
		})
		assert.NoError(t, err)

		// Insert a row with an old, different schedule.
		testutils.MustInsert(t, models.CronJob{
			Queue:        "test-update-cron",
			CronSchedule: "0 30 * * * *",
			NextRunAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		})

		err = p.hydrateCronJobTable()
		assert.NoError(t, err, "hydrate must not return an error")

		var cronJob models.CronJob
		err = db.Model(&cronJob).Where(`"queue" = ?`, "test-update-cron").Select()
		require.NoError(t, err)
		assert.Equal(t, "0 0 * * * *", cronJob.CronSchedule, "schedule must be updated to the newly registered expression")
	})
}

func TestPostgresProcessor_Close(t *testing.T) {
	t.Run("graceful shutdown completes in-flight job before returning", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)

		started := make(chan struct{})

		p := NewPostgresQueue(
			t.Context(),
			clock,
			log,
			config.Configuration{},
			db,
			nil, nil, nil, nil, nil, nil,
		).(*postgresProcessor)

		err := p.Register(t.Context(), "test-slow", func(_ Context, _ []byte) error {
			close(started)
			time.Sleep(1 * time.Second)
			return nil
		})
		assert.NoError(t, err)

		err = p.Start()
		assert.NoError(t, err, "must be able to start the processor")

		err = p.EnqueueAt(t.Context(), "test-slow", clock.Now(), nil)
		assert.NoError(t, err, "must be able to enqueue a job")

		// Wait for the job to start executing.
		select {
		case <-started:
		case <-time.After(10 * time.Second):
			require.Fail(t, "timed out waiting for job to start executing")
		}

		// Close should wait for the in-flight job to finish.
		err = p.Close()
		assert.NoError(t, err, "graceful shutdown must complete without error")

		count, err := db.Model(new(models.Job)).
			Where(`"status" = ?`, models.CompletedJobStatus).
			Count()
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "in-flight job must be completed after graceful shutdown")
	})
}
