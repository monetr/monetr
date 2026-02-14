package background_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
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
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db, log)

		// Create a bogus file record
		file := testutils.MustInsert(t, File{
			AccountId:   bankAccount.AccountId,
			Name:        "sample-part-one.ofx",
			ContentType: models.IntuitQFXContentType,
			Size:        uint64(10),
			CreatedBy:   user.UserId,
		})

		store := mockgen.NewMockStorage(ctrl)
		store.EXPECT().
			Remove(
				gomock.Any(),
				gomock.Eq(file),
			).
			Times(1).
			Return(
				nil,
			)

		job, err := background.NewRemoveFileJob(log, repo, clock, store, background.RemoveFileArguments{
			AccountId: file.AccountId,
			FileId:    file.FileId,
		})
		assert.NoError(t, err, "must be able to create a remove file job")

		err = job.Run(t.Context())
		assert.NoError(t, err, "must be able to remove files")
	})

	t.Run("non-existant file", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db, log)

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

		job, err := background.NewRemoveFileJob(log, repo, clock, store, background.RemoveFileArguments{
			AccountId: bankAccount.AccountId,
			FileId:    "file_bogus",
		})
		assert.NoError(t, err, "must be able to create a remove file job")

		err = job.Run(t.Context())
		assert.EqualError(t, err, "failed to retrieve file record: pg: no rows in result set")
	})
}
