package background

import (
	"context"
	"testing"

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
