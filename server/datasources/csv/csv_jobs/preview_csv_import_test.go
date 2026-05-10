package csv_jobs_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/datasources/csv/csv_jobs"
	"github.com/monetr/monetr/server/datasources/table"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func GetTransactionImportPreview(
	t *testing.T,
	bankAccount BankAccount,
	transactionImportId ID[TransactionImport],
) *TransactionImportPreview {
	db := testutils.GetPgDatabase(t)
	var preview TransactionImportPreview
	err := db.ModelContext(t.Context(), &preview).
		Where(`"account_id" = ?`, bankAccount.AccountId).
		Where(`"bank_account_id" = ?`, bankAccount.BankAccountId).
		Where(`"transaction_import_id" = ?`, transactionImportId).
		Limit(1).
		Select(&preview)
	require.NoError(t, err, "must be able to find transaction import preview")
	require.NotEmpty(
		t,
		preview.TransactionImportPreviewId,
		"must be able to find transaction import preview",
	)
	return &preview
}

func TestPreviewCSVImport(t *testing.T) {
	t.Run("happy path with one matching existing transaction", func(t *testing.T) {
		// monetr's preview job parses the uploaded CSV, looks up which rows
		// already exist via [Transaction.UploadIdentifier], persists a
		// [TransactionImportPreview] with an item per row, and advances the
		// import to the preview status. With the mapping kind=native on
		// "Description" the row's upload identifier ends up being the
		// description text verbatim, which makes seeding a matching transaction
		// easy and predictable.
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ps := pubsub.NewPostgresPubSub(log, db)

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		link := fixtures.GivenIHaveAManualLink(t, clock, user)
		bankAccount := fixtures.GivenIHaveABankAccount(
			t,
			clock,
			&link,
			DepositoryBankAccountType,
			CheckingBankAccountSubType,
		)

		csvData := []byte(
			"Date,Description,Amount\n" +
				"2026-01-15,COFFEE SHOP,-4.50\n" +
				"2026-01-16,GAS STATION,-45.00\n",
		)

		// Seed a transaction whose upload identifier matches one row's
		// Description. The native single-field IDSpec means the row id will
		// equal the column value verbatim.
		existingTransaction := testutils.MustInsert(t, Transaction{
			AccountId:        bankAccount.AccountId,
			BankAccountId:    bankAccount.BankAccountId,
			Amount:           450,
			Date:             clock.Now(),
			Name:             "COFFEE SHOP",
			OriginalName:     "COFFEE SHOP",
			MerchantName:     "COFFEE SHOP",
			OriginalMerchantName: "COFFEE SHOP",
			IsPending:        false,
			Source:           TransactionSourceUpload,
			UploadIdentifier: myownsanity.Pointer("COFFEE SHOP"),
			CreatedAt:        clock.Now(),
		})

		file := testutils.MustInsert(t, File{
			AccountId:   bankAccount.AccountId,
			Name:        "transactions.csv",
			Kind:        "transactions/imports",
			ContentType: models.TextCSVContentType,
			Size:        uint64(len(csvData)),
			CreatedBy:   user.UserId,
			CreatedAt:   clock.Now().UTC(),
		})

		mapping := testutils.MustInsert(t, TransactionImportMapping{
			AccountId: bankAccount.AccountId,
			CreatedBy: user.UserId,
			Mapping: table.Mapping{
				ID: table.IDSpec{
					Kind: table.IDSpecKindNative,
					Fields: []table.FieldRef{
						{
							Name: "Description",
						},
					},
				},
				Amount: table.AmountSpec{
					Kind: table.AmountKindSign,
					Fields: []table.FieldRef{
						{
							Name: "Amount",
						},
					},
				},
				Memo: table.FieldRef{
					Name: "Description",
				},
				Date: table.DateSpec{
					Fields: []table.FieldRef{
						{
							Name: "Date",
						},
					},
					Format: "YYYY-MM-DD",
				},
				Balance: table.BalanceSpec{
					Kind: table.BalanceKindNone,
				},
				Headers: []string{
					"Date",
					"Description",
					"Amount",
				},
			},
		})

		transactionImport := testutils.MustInsert(t, TransactionImport{
			AccountId:                  bankAccount.AccountId,
			BankAccountId:              bankAccount.BankAccountId,
			FileId:                     file.FileId,
			TransactionImportMappingId: &mapping.TransactionImportMappingId,
			Headers: []string{
				"Date",
				"Description",
				"Amount",
			},
			Delimeter: ",",
			Status:    TransactionImportStatusPendingPreview,
			CreatedBy: user.UserId,
		})

		store := mockgen.NewMockStorage(ctrl)
		enqueuer := mockgen.NewMockProcessor(ctrl)

		store.EXPECT().
			Read(
				gomock.Any(),
				gomock.Eq(file),
			).
			Return(
				io.NopCloser(bytes.NewReader(csvData)),
				nil,
			)

		context := mockgen.NewMockContext(ctrl)
		context.EXPECT().Clock().Return(clock).AnyTimes()
		context.EXPECT().DB().Return(db).AnyTimes()
		context.EXPECT().Log().Return(log).AnyTimes()
		context.EXPECT().Enqueuer().Return(enqueuer).AnyTimes()
		context.EXPECT().Publisher().Return(ps).AnyTimes()
		context.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
		context.EXPECT().Storage().Return(store).AnyTimes()

		err := csv_jobs.PreviewCSVImport(
			mockqueue.NewMockContext(context),
			csv_jobs.PreviewCSVImportArguments{
				AccountId:           transactionImport.AccountId,
				BankAccountId:       transactionImport.BankAccountId,
				TransactionImportId: transactionImport.TransactionImportId,
			},
		)
		assert.NoError(t, err, "must be able to preview the csv import")

		// The import row should now be in preview status.
		updatedImport := testutils.MustRetrieve(t, TransactionImport{
			TransactionImportId: transactionImport.TransactionImportId,
			AccountId:           transactionImport.AccountId,
			BankAccountId:       transactionImport.BankAccountId,
		})
		assert.Equal(
			t,
			TransactionImportStatusPreview,
			updatedImport.Status,
			"transaction import should have been moved to preview status",
		)

		// And a preview row should now exist for that import, with one row per
		// CSV line. Rows are expected to come back in the same order they
		// appear in the file, so index by position to cover row-order too.
		preview := GetTransactionImportPreview(t, bankAccount, transactionImport.TransactionImportId)
		require.Len(t, preview.Rows, 2, "preview should contain both csv rows")

		matched := preview.Rows[0]
		unmatched := preview.Rows[1]

		assert.Equal(t, "COFFEE SHOP", matched.Data.ID, "first row should be the COFFEE SHOP entry")
		assert.Equal(t, "GAS STATION", unmatched.Data.ID, "second row should be the GAS STATION entry")

		require.Len(
			t,
			matched.ExistingTransactionIds,
			1,
			"the matching row should have one existing transaction id",
		)
		assert.Equal(
			t,
			existingTransaction.TransactionId,
			matched.ExistingTransactionIds[0],
			"the matching row should reference the seeded transaction",
		)
		assert.Empty(
			t,
			unmatched.ExistingTransactionIds,
			"the non-matching row should not have any existing transaction ids",
		)
	})
}
