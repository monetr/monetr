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
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := jobs.CleanupJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")
	})

	t.Run("removes old jobs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		job := testutils.MustInsert(t, models.Job{
			Queue:       "test queue",
			Signature:   "abc123",
			Input:       "",
			Output:      "",
			Status:      models.PendingJobStatus,
			CreatedAt:   clock.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   clock.Now().Add(-30 * 24 * time.Hour),
			StartedAt:   nil,
			CompletedAt: nil,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := jobs.CleanupJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "removing old jobs should not produce an error")

		testutils.MustDBNotExist(t, job)
	})

	t.Run("will not remove a newer job", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		job := testutils.MustInsert(t, models.Job{
			Queue:       "test queue",
			Signature:   "abc123",
			Input:       "",
			Output:      "",
			Status:      models.PendingJobStatus,
			CreatedAt:   time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-5 * 24 * time.Hour),
			StartedAt:   nil,
			CompletedAt: nil,
		})

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)

		err := jobs.CleanupJobsCron(mockqueue.NewMockContext(context))
		assert.NoError(t, err, "should not return an error when there are no jobs to cleanup")

		testutils.MustDBNotExist(t, job)
	})
}
