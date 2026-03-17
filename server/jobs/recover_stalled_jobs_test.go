package jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/jobs"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRecoverStalledJobsCron(t *testing.T) {
	t.Run("no stalled jobs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.RecoverStalledJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "should not return an error when there are no stalled jobs")
	})

	t.Run("recovers stalled job with retries remaining", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		now := clock.Now()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		startedAt := now.Add(-15 * time.Minute)
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "stalled-retryable",
			Input:     "",
			Output:    "",
			Status:    models.ProcessingJobStatus,
			Attempt:   1,
			CreatedAt: startedAt,
			UpdatedAt: startedAt,
			StartedAt: &startedAt,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.RecoverStalledJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "should recover the stalled job without error")

		updated := testutils.MustDBRead(t, job)
		assert.Equal(t, models.PendingJobStatus, updated.Status, "job should be moved back to pending")
		assert.Equal(t, 2, updated.Attempt, "attempt should be incremented")
		assert.Nil(t, updated.StartedAt, "started_at should be cleared")
	})

	t.Run("marks stalled job as failed when attempts exhausted", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		now := clock.Now()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		startedAt := now.Add(-15 * time.Minute)
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "stalled-exhausted",
			Input:     "",
			Output:    "",
			Status:    models.ProcessingJobStatus,
			Attempt:   5,
			CreatedAt: startedAt,
			UpdatedAt: startedAt,
			StartedAt: &startedAt,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.RecoverStalledJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "should mark exhausted job as failed without error")

		updated := testutils.MustDBRead(t, job)
		assert.Equal(t, models.FailedJobStatus, updated.Status, "job should be marked as failed")
		assert.NotNil(t, updated.CompletedAt, "completed_at should be set")
	})

	t.Run("does not touch jobs within the threshold", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		now := clock.Now()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		// Job started 5 minutes ago, within the 10-minute threshold.
		startedAt := now.Add(-5 * time.Minute)
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "still-running",
			Input:     "",
			Output:    "",
			Status:    models.ProcessingJobStatus,
			Attempt:   1,
			CreatedAt: startedAt,
			UpdatedAt: startedAt,
			StartedAt: &startedAt,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.RecoverStalledJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "should not return an error")

		updated := testutils.MustDBRead(t, job)
		assert.Equal(t, models.ProcessingJobStatus, updated.Status, "job should still be processing")
		assert.Equal(t, 1, updated.Attempt, "attempt should not change")
		assert.NotNil(t, updated.StartedAt, "started_at should not be cleared")
	})

	t.Run("handles mixed stalled jobs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		now := clock.Now()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		stalledAt := now.Add(-15 * time.Minute)
		recentAt := now.Add(-5 * time.Minute)

		// Stalled with retries remaining.
		retryable := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "mixed-retryable",
			Input:     "",
			Output:    "",
			Status:    models.ProcessingJobStatus,
			Attempt:   2,
			CreatedAt: stalledAt,
			UpdatedAt: stalledAt,
			StartedAt: &stalledAt,
		})

		// Stalled with attempts exhausted.
		exhausted := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "mixed-exhausted",
			Input:     "",
			Output:    "",
			Status:    models.ProcessingJobStatus,
			Attempt:   5,
			CreatedAt: stalledAt,
			UpdatedAt: stalledAt,
			StartedAt: &stalledAt,
		})

		// Still running, should not be touched.
		running := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "mixed-running",
			Input:     "",
			Output:    "",
			Status:    models.ProcessingJobStatus,
			Attempt:   1,
			CreatedAt: recentAt,
			UpdatedAt: recentAt,
			StartedAt: &recentAt,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.RecoverStalledJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "should handle mixed stalled jobs without error")

		updatedRetryable := testutils.MustDBRead(t, retryable)
		assert.Equal(t, models.PendingJobStatus, updatedRetryable.Status, "retryable job should be pending")
		assert.Equal(t, 3, updatedRetryable.Attempt, "retryable job attempt should be incremented")
		assert.Nil(t, updatedRetryable.StartedAt, "retryable job started_at should be cleared")

		updatedExhausted := testutils.MustDBRead(t, exhausted)
		assert.Equal(t, models.FailedJobStatus, updatedExhausted.Status, "exhausted job should be failed")
		assert.NotNil(t, updatedExhausted.CompletedAt, "exhausted job completed_at should be set")

		updatedRunning := testutils.MustDBRead(t, running)
		assert.Equal(t, models.ProcessingJobStatus, updatedRunning.Status, "running job should still be processing")
		assert.Equal(t, 1, updatedRunning.Attempt, "running job attempt should not change")
	})
}
