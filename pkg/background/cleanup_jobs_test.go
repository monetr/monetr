package background

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestCleanupJobsJob_Run(t *testing.T) {
	t.Run("no jobs to cleanup", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		handler := NewCleanupJobsHandler(log, db)

		var args interface{}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")
		assert.Equal(t, "no jobs were cleaned up from the jobs table", hook.Entries[len(hook.Entries)-2].Message, "should have no jobs to cleanup")
	})

	t.Run("removes old jobs", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		job := models.Job{
			JobId:      gofakeit.UUID(),
			Name:       "test job",
			Args:       nil,
			EnqueuedAt: time.Now().Add(-30 * 24 * time.Hour),
			StartedAt:  nil,
			FinishedAt: nil,
			Retries:    0,
		}
		_, err := db.Model(&job).Insert(&job)
		assert.NoError(t, err, "must be able to seed the test job")

		handler := NewCleanupJobsHandler(log, db)

		var args interface{}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		lastEntry := hook.Entries[len(hook.Entries)-2]
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")
		assert.Equal(t, "deleted old jobs from the jobs table", lastEntry.Message, "should have no jobs to cleanup")
		assert.Equal(t, 1, lastEntry.Data["deleted"], "should have deleted only one job")

		exists, err := db.Model(&models.Job{}).Where(`"job"."job_id" = ?`, job.JobId).Exists()
		assert.NoError(t, err, "exists query must succeed")
		assert.False(t, exists, "job should not longer exist in the table")
	})

	t.Run("will not remove a newer job", func(t *testing.T) {
		log, hook := testutils.GetTestLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		job := models.Job{
			JobId:      gofakeit.UUID(),
			Name:       "test job",
			Args:       nil,
			EnqueuedAt: time.Now().Add(-5 * 24 * time.Hour),
			StartedAt:  nil,
			FinishedAt: nil,
			Retries:    0,
		}
		_, err := db.Model(&job).Insert(&job)
		assert.NoError(t, err, "must be able to seed the test job")

		handler := NewCleanupJobsHandler(log, db)

		var args interface{}
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), argsEncoded)
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")
		assert.Equal(t, "no jobs were cleaned up from the jobs table", hook.Entries[len(hook.Entries)-2].Message, "should have no jobs to cleanup")

		exists, err := db.Model(&models.Job{}).Where(`"job"."job_id" = ?`, job.JobId).Exists()
		assert.NoError(t, err, "exists query must succeed")
		assert.True(t, exists, "job should not longer exist in the table")
	})
}
