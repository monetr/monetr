package recurring_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const SampleDataPath = "/home/elliotcourant/Downloads/transactions sample.ofx"

func TestSimilarTransactionsBlogPost(t *testing.T) {
	clock := clock.NewMock()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	log := testutils.GetLog(t)
	// Make test quieter
	log.Logger.SetLevel(logrus.ErrorLevel)
	db := testutils.GetPgDatabase(t)
	ps := pubsub.NewPostgresPubSub(log, db)

	user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
	// Force the timezone to be central time
	user.Account.Timezone = "America/Central"
	testutils.MustDBUpdate(t, user.Account)
	link := fixtures.GivenIHaveAPlaidLink(t, clock, user)
	bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

	repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db)
	store := mockgen.NewMockStorage(ctrl)
	enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

	{ // Import our sample data into monetr
		// Read the transaction OFX file
		sampleFileData, err := os.ReadFile(SampleDataPath)
		assert.NoError(t, err, "must be able to read sample file data")

		// Create a bogus file record
		file := testutils.MustInsert(t, File{
			AccountId:   bankAccount.AccountId,
			Name:        "sample.ofx",
			ContentType: string(storage.IntuitQFXContentType),
			Size:        uint64(len(sampleFileData)),
			BlobUri:     "bogus:///bogus",
			CreatedBy:   user.UserId,
		})

		// Create the file upload record for the job.
		upload := testutils.MustInsert(t, TransactionUpload{
			AccountId:     bankAccount.AccountId,
			BankAccountId: bankAccount.BankAccountId,
			FileId:        file.FileId,
			Status:        TransactionUploadStatusPending,
			Error:         nil,
			CreatedBy:     user.UserId,
			ProcessedAt:   nil,
			CompletedAt:   nil,
		})

		// Create the job executor
		job, err := background.NewProcessOFXUploadJob(log, repo, clock, store, ps, enqueuer, background.ProcessOFXUploadArguments{
			AccountId:           upload.AccountId,
			BankAccountId:       upload.BankAccountId,
			TransactionUploadId: upload.TransactionUploadId,
		})
		assert.NoError(t, err, "must be able to create an OFX upload job")

		{ // Mock out our expected calls from within the job
			store.EXPECT().
				Read(
					gomock.Any(),
					gomock.Eq(file.BlobUri),
				).
				Return(
					io.NopCloser(bytes.NewReader(sampleFileData)),
					storage.IntuitQFXContentType,
					nil,
				)

			enqueuer.EXPECT().
				EnqueueJob(
					gomock.Any(),
					background.CalculateTransactionClusters,
					gomock.Eq(background.CalculateTransactionClustersArguments{
						AccountId:     bankAccount.AccountId,
						BankAccountId: bankAccount.BankAccountId,
					}),
				)
			enqueuer.EXPECT().
				EnqueueJob(
					gomock.Any(),
					background.RemoveFile,
					gomock.Eq(background.RemoveFileArguments{
						AccountId: file.AccountId,
						FileId:    file.FileId,
					}),
				)
		}

		// Run our import job
		err = job.Run(t.Context())
		assert.NoError(t, err, "must be able to import ofx transactions for sample")
	}

	// We should now have a bunch of transactions, but now we need to actually run
	// the transaction cluster code.

	// Make test louder again
	log.Logger.SetLevel(logrus.TraceLevel)

	{ // Calculate our similar transactions
		job, err := background.NewCalculateTransactionClustersJob(log, db, clock, background.CalculateTransactionClustersArguments{
			AccountId:     bankAccount.AccountId,
			BankAccountId: bankAccount.BankAccountId,
		})
		assert.NoError(t, err, "must be able to create transaction cluster job")

		err = job.Run(t.Context())
		assert.NoError(t, err, "must be able to calculate similar transactions")
	}

}
