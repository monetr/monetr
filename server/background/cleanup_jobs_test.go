package background

import (
	"context"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestCleanupJobsJob_Run(t *testing.T) {
	t.Run("no jobs to cleanup", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		handler := NewCleanupJobsHandler(log, db)

		var args any
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), log, argsEncoded)
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")
	})

	t.Run("removes old jobs", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		job := models.Job{
			Queue:       "test queue",
			Signature:   "abc123",
			Input:       "",
			Output:      "",
			Status:      models.PendingJobStatus,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			StartedAt:   nil,
			CompletedAt: nil,
		}
		_, err := db.Model(&job).Insert(&job)
		assert.NoError(t, err, "must be able to seed the test job")

		handler := NewCleanupJobsHandler(log, db)

		var args any
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), log, argsEncoded)
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")

		exists, err := db.Model(&models.Job{}).Where(`"job"."job_id" = ?`, job.JobId).Exists()
		assert.NoError(t, err, "exists query must succeed")
		assert.False(t, exists, "job should not longer exist in the table")
	})

	t.Run("will not remove a newer job", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		job := models.Job{
			Queue:       "test queue",
			Signature:   "abc123",
			Input:       "",
			Output:      "",
			Status:      models.PendingJobStatus,
			CreatedAt:   time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-5 * 24 * time.Hour),
			StartedAt:   nil,
			CompletedAt: nil,
		}
		_, err := db.Model(&job).Insert(&job)
		assert.NoError(t, err, "must be able to seed the test job")

		handler := NewCleanupJobsHandler(log, db)

		var args any
		argsEncoded, err := DefaultJobMarshaller(args)
		assert.NoError(t, err, "must be able to marshal arguments")

		err = handler.HandleConsumeJob(context.Background(), log, argsEncoded)
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")

		exists, err := db.Model(&models.Job{}).Where(`"job"."job_id" = ?`, job.JobId).Exists()
		assert.NoError(t, err, "exists query must succeed")
		assert.True(t, exists, "job should not longer exist in the table")
	})
}
