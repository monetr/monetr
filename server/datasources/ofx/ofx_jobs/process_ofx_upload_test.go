package ofx_jobs_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/datasources/ofx/ofx_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/similar"
	"github.com/monetr/monetr/server/storage/storage_jobs"
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
		user.Account.Timezone = "America/Chicago"
		testutils.MustDBUpdate(t, user.Account)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)

		store := mockgen.NewMockStorage(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		{ // Import our sample data into monetr
			// Read the transaction OFX file
			sampleFileData := fixtures.LoadFile(t, "sample-part-one.ofx")

			// Create a bogus file record
			file := testutils.MustInsert(t, File{
				AccountId:   bankAccount.AccountId,
				Name:        "sample-part-one.ofx",
				Kind:        "transactions/uploads",
				ContentType: models.IntuitQFXContentType,
				Size:        uint64(len(sampleFileData)),
				CreatedBy:   user.UserId,
				CreatedAt:   clock.Now().UTC(),
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

			{ // Mock out our expected calls from within the job
				store.EXPECT().
					Read(
						gomock.Any(),
						gomock.Eq(file),
					).
					Return(
						io.NopCloser(bytes.NewReader(sampleFileData)),
						nil,
					)
				enqueuer.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(similar.CalculateTransactionClusters),
						gomock.Any(),
						gomock.Eq(similar.CalculateTransactionClustersArguments{
							AccountId:     bankAccount.AccountId,
							BankAccountId: bankAccount.BankAccountId,
						}),
					).
					Return(nil).
					Times(1)
				enqueuer.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(storage_jobs.RemoveFile),
						gomock.Any(),
						gomock.Eq(storage_jobs.RemoveFileArguments{
							AccountId: file.AccountId,
							FileId:    file.FileId,
						}),
					).
					Return(nil).
					Times(1)
			}

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Publisher().Return(ps).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
			context.EXPECT().Storage().Return(store).AnyTimes()

			// Run our import job
			err := ofx_jobs.ProcessOFXUpload(
				mockqueue.NewMockContext(context),
				ofx_jobs.ProcessOFXUploadArguments{
					AccountId:           upload.AccountId,
					BankAccountId:       upload.BankAccountId,
					TransactionUploadId: upload.TransactionUploadId,
				},
			)
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
				Kind:        "transactions/uploads",
				ContentType: models.IntuitQFXContentType,
				Size:        uint64(len(sampleFileData)),
				CreatedBy:   user.UserId,
				CreatedAt:   clock.Now().UTC(),
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

			{ // Mock out our expected calls from within the job
				store.EXPECT().
					Read(
						gomock.Any(),
						gomock.Eq(file),
					).
					Return(
						io.NopCloser(bytes.NewReader(sampleFileData)),
						nil,
					)
				enqueuer.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(similar.CalculateTransactionClusters),
						gomock.Any(),
						gomock.Eq(similar.CalculateTransactionClustersArguments{
							AccountId:     bankAccount.AccountId,
							BankAccountId: bankAccount.BankAccountId,
						}),
					).
					Return(nil).
					Times(1)
				enqueuer.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(storage_jobs.RemoveFile),
						gomock.Any(),
						gomock.Eq(storage_jobs.RemoveFileArguments{
							AccountId: file.AccountId,
							FileId:    file.FileId,
						}),
					).
					Return(nil).
					Times(1)
			}

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
			context.EXPECT().Publisher().Return(ps).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
			context.EXPECT().Storage().Return(store).AnyTimes()

			// Run our import job
			err := ofx_jobs.ProcessOFXUpload(
				mockqueue.NewMockContext(context),
				ofx_jobs.ProcessOFXUploadArguments{
					AccountId:           upload.AccountId,
					BankAccountId:       upload.BankAccountId,
					TransactionUploadId: upload.TransactionUploadId,
				},
			)
			assert.NoError(t, err, "must be able to import ofx transactions for sample")
			assert.NotNil(t, GetTxnByUploadIdentifier(t, bankAccount, "8a34b9c89506be8c0195711720250751"))
			// Find the new transaction from the second file
			assert.NotNil(t, GetTxnByUploadIdentifier(t, bankAccount, "8a348ece9588213401958a84627458ab"))
		}
	})

	t.Run("valid file no transactions", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ps := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		// Force the timezone to be central time
		user.Account.Timezone = "America/Chicago"
		testutils.MustDBUpdate(t, user.Account)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		store := mockgen.NewMockStorage(ctrl)
		processor := mockgen.NewMockProcessor(ctrl)

		{ // Import our sample data into monetr
			// Read the transaction OFX file
			sampleFileData := fixtures.LoadFile(t, "sample-no-txn.ofx")

			// Create a bogus file record
			file := testutils.MustInsert(t, File{
				AccountId:   bankAccount.AccountId,
				Name:        "sample-part-one.ofx",
				Kind:        "transactions/uploads",
				ContentType: models.IntuitQFXContentType,
				Size:        uint64(len(sampleFileData)),
				CreatedBy:   user.UserId,
				CreatedAt:   clock.Now().UTC(),
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

			{ // Mock out our expected calls from within the job
				store.EXPECT().
					Read(
						gomock.Any(),
						gomock.Eq(file),
					).
					Return(
						io.NopCloser(bytes.NewReader(sampleFileData)),
						nil,
					)

				processor.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(similar.CalculateTransactionClusters),
						gomock.Any(),
						gomock.Eq(similar.CalculateTransactionClustersArguments{
							AccountId:     bankAccount.AccountId,
							BankAccountId: bankAccount.BankAccountId,
						}),
					).
					Return(nil).
					Times(1)
				processor.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(storage_jobs.RemoveFile),
						gomock.Any(),
						gomock.Eq(storage_jobs.RemoveFileArguments{
							AccountId: file.AccountId,
							FileId:    file.FileId,
						}),
					).
					Return(nil).
					Times(1)
			}
			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Enqueuer().Return(processor).AnyTimes()
			context.EXPECT().Publisher().Return(ps).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
			context.EXPECT().Storage().Return(store).AnyTimes()

			// Run our import job
			err := ofx_jobs.ProcessOFXUpload(
				mockqueue.NewMockContext(context),
				ofx_jobs.ProcessOFXUploadArguments{
					AccountId:           upload.AccountId,
					BankAccountId:       upload.BankAccountId,
					TransactionUploadId: upload.TransactionUploadId,
				},
			)
			assert.NoError(t, err, "must be able to import ofx transactions for sample")
			fixtures.AssertThatIHaveZeroTransactions(t, user.AccountId)
		}
	})

	t.Run("valid file, no amount on one transaction", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ps := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		// Force the timezone to be central time
		user.Account.Timezone = "America/Chicago"
		testutils.MustDBUpdate(t, user.Account)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		store := mockgen.NewMockStorage(ctrl)
		processor := mockgen.NewMockProcessor(ctrl)

		{ // Import our sample data into monetr
			// Read the transaction OFX file
			sampleFileData := fixtures.LoadFile(t, "sample-no-amount.ofx")

			// Create a bogus file record
			file := testutils.MustInsert(t, File{
				AccountId:   bankAccount.AccountId,
				Name:        "sample-no-amount.ofx",
				Kind:        "transactions/uploads",
				ContentType: models.IntuitQFXContentType,
				Size:        uint64(len(sampleFileData)),
				CreatedBy:   user.UserId,
				CreatedAt:   clock.Now().UTC(),
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

			{ // Mock out our expected calls from within the job
				store.EXPECT().
					Read(
						gomock.Any(),
						gomock.Eq(file),
					).
					Return(
						io.NopCloser(bytes.NewReader(sampleFileData)),
						nil,
					)
				processor.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(similar.CalculateTransactionClusters),
						gomock.Any(),
						gomock.Eq(similar.CalculateTransactionClustersArguments{
							AccountId:     bankAccount.AccountId,
							BankAccountId: bankAccount.BankAccountId,
						}),
					).
					Return(nil).
					Times(1)
				processor.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(storage_jobs.RemoveFile),
						gomock.Any(),
						gomock.Eq(storage_jobs.RemoveFileArguments{
							AccountId: file.AccountId,
							FileId:    file.FileId,
						}),
					).
					Return(nil).
					Times(1)
			}

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Enqueuer().Return(processor).AnyTimes()
			context.EXPECT().Publisher().Return(ps).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
			context.EXPECT().Storage().Return(store).AnyTimes()

			// Run our import job
			err := ofx_jobs.ProcessOFXUpload(
				mockqueue.NewMockContext(context),
				ofx_jobs.ProcessOFXUploadArguments{
					AccountId:           upload.AccountId,
					BankAccountId:       upload.BankAccountId,
					TransactionUploadId: upload.TransactionUploadId,
				},
			)
			assert.NoError(t, err, "must be able to import ofx transactions for sample")

			{ // Check on our "blank" transaction
				txn := GetTxnByUploadIdentifier(t, bankAccount, "8a34b9c89506be8c0195711720250751")
				assert.NotNil(t, txn)
				assert.EqualValues(t, 0, txn.Amount, "amount should be 0 because it was blank in the file")
			}

			{ // Check on the transaction that had an amount
				txn := GetTxnByUploadIdentifier(t, bankAccount, "8a34b9c89506be8c0195711720250753")
				assert.NotNil(t, txn)
				assert.EqualValues(t, 3216, txn.Amount, "amount should be 0 because it was blank in the file")
			}
		}
	})

	t.Run("handle missing curdef in ofx file", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ps := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		// Force the timezone to be central time
		user.Account.Timezone = "America/Chicago"
		testutils.MustDBUpdate(t, user.Account)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(t, clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		{ // Update the bank account to be MXN currency
			bankAccount.Currency = "MXN"
			testutils.MustDBUpdate(t, &bankAccount)
		}

		store := mockgen.NewMockStorage(ctrl)
		processor := mockgen.NewMockProcessor(ctrl)

		{ // Import our sample data into monetr
			// Read the transaction OFX file
			sampleFileData := fixtures.LoadFile(t, "no-curdef-mxn.ofx")

			// Create a bogus file record
			file := testutils.MustInsert(t, File{
				AccountId:   bankAccount.AccountId,
				Name:        "no-curdef-mxn.ofx",
				Kind:        "transactions/uploads",
				ContentType: models.IntuitQFXContentType,
				Size:        uint64(len(sampleFileData)),
				CreatedBy:   user.UserId,
				CreatedAt:   clock.Now().UTC(),
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

			{ // Mock out our expected calls from within the job
				store.EXPECT().
					Read(
						gomock.Any(),
						gomock.Eq(file),
					).
					Return(
						io.NopCloser(bytes.NewReader(sampleFileData)),
						nil,
					)
				processor.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(similar.CalculateTransactionClusters),
						gomock.Any(),
						gomock.Eq(similar.CalculateTransactionClustersArguments{
							AccountId:     bankAccount.AccountId,
							BankAccountId: bankAccount.BankAccountId,
						}),
					).
					Return(nil).
					Times(1)
				processor.EXPECT().
					EnqueueAt(
						gomock.Any(),
						mockqueue.EqQueue(storage_jobs.RemoveFile),
						gomock.Any(),
						gomock.Eq(storage_jobs.RemoveFileArguments{
							AccountId: file.AccountId,
							FileId:    file.FileId,
						}),
					).
					Return(nil).
					Times(1)
			}

			context := mockgen.NewMockContext(ctrl)
			context.EXPECT().Clock().Return(clock).AnyTimes()
			context.EXPECT().DB().Return(db).AnyTimes()
			context.EXPECT().Log().Return(log).AnyTimes()
			context.EXPECT().Enqueuer().Return(processor).AnyTimes()
			context.EXPECT().Publisher().Return(ps).AnyTimes()
			context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
			context.EXPECT().Storage().Return(store).AnyTimes()

			// Run our import job
			err := ofx_jobs.ProcessOFXUpload(
				mockqueue.NewMockContext(context),
				ofx_jobs.ProcessOFXUploadArguments{
					AccountId:           upload.AccountId,
					BankAccountId:       upload.BankAccountId,
					TransactionUploadId: upload.TransactionUploadId,
				},
			)
			assert.NoError(t, err, "must be able to import ofx transactions for sample")
			assert.NotNil(t, GetTxnByUploadIdentifier(t, bankAccount, "db0676765969665a5a920576638051aee50fd0ca"))
		}
	})
}
