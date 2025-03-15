package background_test

import (
	"bytes"
	"io"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func GetTxnByUploadIdentifier(t *testing.T, bankAccount BankAccount, uploadId string) *Transaction {
	db := testutils.GetPgDatabase(t)
	var txn Transaction
	err := db.ModelContext(t.Context(), &txn).
		Where(`"account_id" = ?`, bankAccount.AccountId).
		Where(`"bank_account_id" = ?`, bankAccount.BankAccountId).
		Where(`"upload_identifier" = ?`, uploadId).
		Limit(1).
		Select(&txn)
	require.NoError(t, err, "must be able to find transaction by upload identifier: %s", uploadId)
	require.NotEmpty(t, txn.TransactionId, "must be able to find transaction by upload identifier: %s", uploadId)
	return &txn
}

func TestProcessOFXUploadJob_Run(t *testing.T) {
	t.Run("valid file two part upload", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ps := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		// Force the timezone to be central time
		user.Account.Timezone = "America/Central"
		testutils.MustDBUpdate(t, user.Account)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		repo := repository.NewRepositoryFromSession(clock, user.UserId, user.AccountId, db, log)
		store := mockgen.NewMockStorage(ctrl)
		enqueuer := mockgen.NewMockJobEnqueuer(ctrl)

		{ // Import our sample data into monetr
			// Read the transaction OFX file
			sampleFileData := fixtures.LoadFile(t, "sample-part-one.ofx")

			// Create a bogus file record
			file := testutils.MustInsert(t, File{
				AccountId:   bankAccount.AccountId,
				Name:        "sample-part-one.ofx",
				ContentType: string(storage.IntuitQFXContentType),
				Size:        uint64(len(sampleFileData)),
				BlobUri:     "bogus:///bogus-part-one",
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
			assert.NotNil(t, GetTxnByUploadIdentifier(t, bankAccount, "8a34b9c89506be8c0195711720250751"))
		}

		{ // Import our sample data into monetr
			// Read the transaction OFX file
			sampleFileData := fixtures.LoadFile(t, "sample-part-two.ofx")

			// Create a bogus file record
			file := testutils.MustInsert(t, File{
				AccountId:   bankAccount.AccountId,
				Name:        "sample-part-two.ofx",
				ContentType: string(storage.IntuitQFXContentType),
				Size:        uint64(len(sampleFileData)),
				BlobUri:     "bogus:///bogus-part-two",
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
			assert.NotNil(t, GetTxnByUploadIdentifier(t, bankAccount, "8a34b9c89506be8c0195711720250751"))
			// Find the new transaction from the second file
			assert.NotNil(t, GetTxnByUploadIdentifier(t, bankAccount, "8a348ece9588213401958a84627458ab"))
		}
	})
}
