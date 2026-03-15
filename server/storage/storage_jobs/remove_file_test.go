package storage_jobs_test

import (
	"testing"

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

func TestRemoveFileJob_Run(t *testing.T) {
	t.Run("remove file from storage", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		// Create a bogus file record
		file := testutils.MustInsert(t, models.File{
			AccountId:   bankAccount.AccountId,
			Name:        "sample-part-one.ofx",
			Kind:        "transactions/uploads",
			ContentType: models.IntuitQFXContentType,
			Size:        uint64(10),
			CreatedBy:   user.UserId,
			CreatedAt:   clock.Now().UTC(),
		})

		store := mockgen.NewMockStorage(ctrl)
		store.EXPECT().
			Remove(
				gomock.Any(),
				testutils.NewGenericMatcher(func(input models.File) bool {
					return myownsanity.Every(
						assert.Equal(t, file.FileId, input.FileId),
						assert.Equal(t, file.AccountId, input.AccountId),
						assert.NotNil(t, input.DeletedAt),
					)
				}),
			).
			Times(1).
			Return(
				nil,
			)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Storage().Return(store).AnyTimes()

		err := storage_jobs.RemoveFile(
			mockqueue.NewMockContext(context),
			storage_jobs.RemoveFileArguments{
				AccountId: file.AccountId,
				FileId:    file.FileId,
			},
		)
		assert.NoError(t, err, "must be able to run the remove file job")
	})

	t.Run("non-existant file", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			models.DepositoryBankAccountType,
			models.CheckingBankAccountSubType,
		)

		store := mockgen.NewMockStorage(ctrl)
		store.EXPECT().
			Remove(
				gomock.Any(),
				gomock.Any(),
			).
			Times(0).
			Return(
				nil,
			)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Storage().Return(store).AnyTimes()

		err := storage_jobs.RemoveFile(
			mockqueue.NewMockContext(context),
			storage_jobs.RemoveFileArguments{
				AccountId: bankAccount.AccountId,
				FileId:    "file_bogus",
			},
		)
		assert.EqualError(t, err, "failed to retrieve file record: pg: no rows in result set")
	})
}
