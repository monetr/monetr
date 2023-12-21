package background

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestPostgresJobProcessor_RegisterJob(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)
		configuration := config.BackgroundJobs{
			Engine:      config.BackgroundJobEnginePostgreSQL,
			Scheduler:   config.BackgroundJobSchedulerInternal,
			JobSchedule: map[string]string{},
		}

		processor := NewPostgresJobProcessor(log, configuration, clock, db, nil)

		testHandler := NewTestJobHandler(
			t,
			func(
				t *testing.T,
				ctx context.Context,
				data []byte,
			) error {
				// No-Op
				return nil
			},
		)

		// If the processor isnt started and the job has not already been registered
		// we should be able to register the job without error.
		err := processor.RegisterJob(context.Background(), testHandler)
		assert.NoError(t, err)
	})

	t.Run("cant register duplicates", func(t *testing.T) {
		clock := clock.NewMock()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)
		configuration := config.BackgroundJobs{
			Engine:      config.BackgroundJobEnginePostgreSQL,
			Scheduler:   config.BackgroundJobSchedulerInternal,
			JobSchedule: map[string]string{},
		}

		processor := NewPostgresJobProcessor(log, configuration, clock, db, nil)

		testHandler := NewTestJobHandler(
			t,
			func(
				t *testing.T,
				ctx context.Context,
				data []byte,
			) error {
				// No-Op
				return nil
			},
		)

		// If the processor isnt started and the job has not already been registered
		// we should be able to register the job without error.
		err := processor.RegisterJob(context.Background(), testHandler)
		assert.NoError(t, err)

		// But if we register the job again then it should fail
		err = processor.RegisterJob(context.Background(), testHandler)
		assert.Error(t, err, "should return an error if the job is already registered")
	})
}

func TestPostgresJobProcessor_CronJobs(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.New()
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		log := testutils.GetLog(t)
		configuration := config.BackgroundJobs{
			Engine:      config.BackgroundJobEnginePostgreSQL,
			Scheduler:   config.BackgroundJobSchedulerInternal,
			JobSchedule: map[string]string{},
		}

		enqueuer := NewPostgresJobEnqueuer(log, db, clock)
		processor := NewPostgresJobProcessor(log, configuration, clock, db, enqueuer)

		var counter int32
		testCronHandler := NewTestCronJobHandler(
			t,
			"* * * * * *",
			func(_ *testing.T, _ context.Context, _ []byte) error {
				atomic.AddInt32(&counter, 1)
				return nil
			},
		)

		// If the processor isnt started and the job has not already been registered
		// we should be able to register the job without error.
		err := processor.RegisterJob(context.Background(), testCronHandler)
		assert.NoError(t, err)

		err = processor.Start()
		assert.NoError(t, err, "should be able to star the processor")
		defer processor.Close()

		time.Sleep(2 * time.Second)

		// After 2 seconds make sure the counter is greater than 0, we should have processed the cron
		assert.Greater(t, atomic.LoadInt32(&counter), int32(0))
	})
}
