package storage_jobs_test

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage/storage_jobs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCleanupFilesCron(t *testing.T) {
	t.Run("no expired files", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			context.EXPECT().Log().Return(log).AnyTimes()

			err := storage_jobs.CleanupFilesCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("enqueues expired files for removal", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		// Create a file that has already expired and has not been reconciled.
		expiredFile := testutils.MustInsert(t, models.File{
			AccountId:   user.AccountId,
			Name:        "expired-upload.ofx",
			Kind:        "transactions/uploads",
			ContentType: models.IntuitQFXContentType,
			Size:        uint64(100),
			CreatedBy:   user.UserId,
			CreatedAt:   clock.Now().UTC(),
			ExpiresAt:   myownsanity.Pointer(clock.Now().Add(-1 * time.Hour)),
		})

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(storage_jobs.RemoveFile),
				gomock.Any(),
				gomock.Eq(storage_jobs.RemoveFileArguments{
					AccountId: expiredFile.AccountId,
					FileId:    expiredFile.FileId,
				}),
			).
			Return(nil).
			Times(1)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			err := storage_jobs.CleanupFilesCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("does not enqueue files that have not expired", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		// Create a file that expires in the future.
		testutils.MustInsert(t, models.File{
			AccountId:   user.AccountId,
			Name:        "not-yet-expired.ofx",
			Kind:        "transactions/uploads",
			ContentType: models.IntuitQFXContentType,
			Size:        uint64(100),
			CreatedBy:   user.UserId,
			CreatedAt:   clock.Now().UTC(),
			ExpiresAt:   myownsanity.Pointer(clock.Now().Add(24 * time.Hour)),
		})

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			context.EXPECT().Log().Return(log).AnyTimes()

			err := storage_jobs.CleanupFilesCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("does not enqueue already reconciled files", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		// Create an expired file that has already been reconciled.
		testutils.MustInsert(t, models.File{
			AccountId:    user.AccountId,
			Name:         "already-reconciled.ofx",
			Kind:         "transactions/uploads",
			ContentType:  models.IntuitQFXContentType,
			Size:         uint64(100),
			CreatedBy:    user.UserId,
			CreatedAt:    clock.Now().UTC(),
			ExpiresAt:    myownsanity.Pointer(clock.Now().Add(-1 * time.Hour)),
			ReconciledAt: myownsanity.Pointer(clock.Now()),
		})

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Return(nil).
			Times(0)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).Times(0)
			context.EXPECT().Log().Return(log).AnyTimes()

			err := storage_jobs.CleanupFilesCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})

	t.Run("enqueues only expired unreconciled files", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)

		// Create an expired, unreconciled file -- should be enqueued.
		expiredFile := testutils.MustInsert(t, models.File{
			AccountId:   user.AccountId,
			Name:        "expired-unreconciled.ofx",
			Kind:        "transactions/uploads",
			ContentType: models.IntuitQFXContentType,
			Size:        uint64(100),
			CreatedBy:   user.UserId,
			CreatedAt:   clock.Now().UTC(),
			ExpiresAt:   myownsanity.Pointer(clock.Now().Add(-1 * time.Hour)),
		})

		// Create a file that has not expired -- should NOT be enqueued.
		testutils.MustInsert(t, models.File{
			AccountId:   user.AccountId,
			Name:        "still-valid.ofx",
			Kind:        "transactions/uploads",
			ContentType: models.IntuitQFXContentType,
			Size:        uint64(100),
			CreatedBy:   user.UserId,
			CreatedAt:   clock.Now().UTC(),
			ExpiresAt:   myownsanity.Pointer(clock.Now().Add(24 * time.Hour)),
		})

		// Create an expired but already reconciled file -- should NOT be enqueued.
		testutils.MustInsert(t, models.File{
			AccountId:    user.AccountId,
			Name:         "expired-reconciled.ofx",
			Kind:         "transactions/uploads",
			ContentType:  models.IntuitQFXContentType,
			Size:         uint64(100),
			CreatedBy:    user.UserId,
			CreatedAt:    clock.Now().UTC(),
			ExpiresAt:    myownsanity.Pointer(clock.Now().Add(-2 * time.Hour)),
			ReconciledAt: myownsanity.Pointer(clock.Now()),
		})

		enqueuer := mockgen.NewMockProcessor(ctrl)
		enqueuer.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(storage_jobs.RemoveFile),
				gomock.Any(),
				gomock.Eq(storage_jobs.RemoveFileArguments{
					AccountId: expiredFile.AccountId,
					FileId:    expiredFile.FileId,
				}),
			).
			Return(nil).
			Times(1)

		{
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()

			err := storage_jobs.CleanupFilesCron(
				mockqueue.NewMockContext(context),
			)
			assert.NoError(t, err)
		}
	})
}
