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

func TestCleanupJobsJob_Run(t *testing.T) {
	t.Run("no jobs to cleanup", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.CleanupJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")
	})

	t.Run("removes old jobs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		now := clock.Now()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		job := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "abc123",
			Input:     "",
			Output:    "",
			Status:    models.PendingJobStatus,
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-30 * 24 * time.Hour),
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.CleanupJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "removing old jobs should not produce an error")

		testutils.MustDBNotExist(t, job)
	})

	t.Run("will not remove a pending job with future priority", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		now := clock.Now()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		// Old pending job but with a future priority (scheduled to run later).
		// Should NOT be deleted even though created_at is past the 15 day cutoff.
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "future-pending",
			Input:     "",
			Output:    "",
			Status:    models.PendingJobStatus,
			Priority:  uint64(now.Add(24 * time.Hour).Unix()),
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-30 * 24 * time.Hour),
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.CleanupJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err)

		testutils.MustDBExist(t, job)
	})

	t.Run("removes old pending job with past priority", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		clock.Set(time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC))
		now := clock.Now()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		// Old pending job with a past priority. This should be cleaned up.
		job := testutils.MustInsert(t, models.Job{
			Queue:     "test queue",
			Signature: "old-pending",
			Input:     "",
			Output:    "",
			Status:    models.PendingJobStatus,
			Priority:  uint64(now.Add(-20 * 24 * time.Hour).Unix()),
			CreatedAt: now.Add(-30 * 24 * time.Hour),
			UpdatedAt: now.Add(-30 * 24 * time.Hour),
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()

		err := jobs.CleanupJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err)

		testutils.MustDBNotExist(t, job)
	})
}
