package controller_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/monetr/monetr/server/datasources/csv/csv_jobs"
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

func TestGetTransactionImportPreviewByTransactionImportId(t *testing.T) {
	t.Run("upload, map, preview, fetch", func(t *testing.T) {
		// End-to-end happy path for the CSV import preview flow. The user uploads a
		// CSV, creates a mapping that lines up with the file's headers, PATCHes the
		// import to the pending-preview status (which enqueues the preview job in
		// the same db transaction as the patch), the job runs, and then the preview
		// can be fetched via the API. monetr's queue is mocked in tests so the
		// enqueue is observed but the job is run inline here using the captured
		// arguments and a fresh [queue.Context] backed by
		// [mockqueue.NewMockContext].
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var existingTransaction Transaction

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)

			// Seed a transaction whose upload identifier matches one of the CSV rows.
			// The mapping uses kind=native on the Description column so the row's id
			// will equal the description string verbatim, which is what gets compared
			// against this field.
			existingTransaction = testutils.MustInsert(t, Transaction{
				AccountId:            bank.AccountId,
				BankAccountId:        bank.BankAccountId,
				Amount:               450,
				Date:                 app.Clock.Now(),
				Name:                 "COFFEE SHOP",
				OriginalName:         "COFFEE SHOP",
				MerchantName:         "COFFEE SHOP",
				OriginalMerchantName: "COFFEE SHOP",
				IsPending:            false,
				Source:               TransactionSourceUpload,
				UploadIdentifier:     myownsanity.Pointer("COFFEE SHOP"),
				CreatedAt:            app.Clock.Now(),
			})
		}

		csvData := []byte(
			"Date,Description,Amount\n" +
				"2026-01-15,COFFEE SHOP,-4.50\n" +
				"2026-01-16,GAS STATION,-45.00\n",
		)

		// Capture the bytes flowing through Store so the same content can be
		// returned from Read when the job runs against storage.
		var stored []byte
		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(1).
			DoAndReturn(func(_ context.Context, body io.Reader, _ models.File) error {
				b, err := io.ReadAll(body)
				require.NoError(t, err, "must be able to read upload bytes")
				stored = b
				return nil
			})

		var transactionImportId ID[TransactionImport]
		{ // Upload the CSV. This creates the import in mapping status.
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
				WithPath("bankAccountId", bank.BankAccountId).
				WithMultipart().
				WithFileBytes("data", "transactions.csv", csvData).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.status").String().IsEqual("mapping")
			response.JSON().Path("$.headers").Array().IsEqual([]string{"Date", "Description", "Amount"})
			transactionImportId = ID[TransactionImport](response.JSON().Path("$.transactionImportId").String().Raw())
		}

		var mappingId ID[TransactionImportMapping]
		{ // Create a mapping whose headers line up with the uploaded CSV. monetr
			// derives the row's upload identifier from the IDSpec, so a kind=native
			// single-field mapping makes the identifier equal to the column value
			// verbatim (here, the Description text).
			response := e.POST("/api/mappings").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"mapping": map[string]any{
						"id": map[string]any{
							"kind": "native",
							"fields": []any{
								map[string]any{
									"name": "Description",
								},
							},
						},
						"amount": map[string]any{
							"kind": "sign",
							"fields": []any{
								map[string]any{
									"name": "Amount",
								},
							},
						},
						"memo": map[string]any{
							"name": "Description",
						},
						"date": map[string]any{
							"fields": []any{
								map[string]any{
									"name": "Date",
								},
							},
							"format": "YYYY-MM-DD",
						},
						"balance": map[string]any{
							"kind": "none",
						},
						"headers": []string{
							"Date",
							"Description",
							"Amount",
						},
					},
				}).
				Expect()

			response.Status(http.StatusOK)
			mappingId = ID[TransactionImportMapping](response.JSON().Path("$.transactionImportMappingId").String().Raw())
		}

		// The PATCH wraps the enqueue in the same db transaction as the status
		// update via [controller.enqueueJob], so WithTransaction lands first and
		// the EnqueueAt call lands on the transactional queue. Capture the args
		// here so the job can be invoked inline below.
		var jobArgs csv_jobs.PreviewCSVImportArguments
		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(csv_jobs.PreviewCSVImport),
				gomock.Any(),
				testutils.NewGenericMatcher(func(args csv_jobs.PreviewCSVImportArguments) bool {
					return myownsanity.Every(
						assert.EqualValues(t, bank.AccountId, args.AccountId, "Account ID should match"),
						assert.EqualValues(t, bank.BankAccountId, args.BankAccountId, "Bank Account ID should match"),
						assert.EqualValues(t, transactionImportId, args.TransactionImportId, "Transaction Import ID should match"),
					)
				}),
			).
			Times(1).
			DoAndReturn(func(_ context.Context, _ string, _ time.Time, args csv_jobs.PreviewCSVImportArguments) error {
				jobArgs = args
				return nil
			})

		{ // Move the import to pending-preview, which enqueues the job.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionImportId", transactionImportId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"transactionImportMappingId": mappingId,
					"status":                     TransactionImportStatusPendingPreview,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transactionImportId").IsEqual(transactionImportId)
			response.JSON().Path("$.transactionImportMappingId").IsEqual(mappingId)
			response.JSON().Path("$.status").String().IsEqual("pending-preview")
		}

		{ // Bridge the gap between the mocked enqueue and the GET. The queue mock
			// does not run jobs, so invoke [csv_jobs.PreviewCSVImport] directly with
			// the captured args. Mirrors the process_ofx_upload_test.go pattern; uses
			// a fresh gomock controller for the job context's expectations.
			ctrl := gomock.NewController(t)
			log := testutils.GetLog(t)
			db := testutils.GetPgDatabase(t)
			ps := pubsub.NewPostgresPubSub(log, db)

			app.Storage.EXPECT().
				Read(
					gomock.Any(),
					gomock.Any(),
				).
				Times(1).
				Return(io.NopCloser(bytes.NewReader(stored)), nil)

			jobCtx := mockgen.NewMockContext(ctrl)
			jobCtx.EXPECT().Clock().Return(app.Clock).AnyTimes()
			jobCtx.EXPECT().DB().Return(db).AnyTimes()
			jobCtx.EXPECT().Log().Return(log).AnyTimes()
			jobCtx.EXPECT().Enqueuer().Return(app.Queue).AnyTimes()
			jobCtx.EXPECT().Publisher().Return(ps).AnyTimes()
			jobCtx.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).Times(1)
			jobCtx.EXPECT().Storage().Return(app.Storage).AnyTimes()

			err := csv_jobs.PreviewCSVImport(
				mockqueue.NewMockContext(jobCtx),
				jobArgs,
			)
			require.NoError(t, err, "must be able to run the preview job")
		}

		{ // Now the preview is persisted, fetch it through the API.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}/preview").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionImportId", transactionImportId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transactionImportId").IsEqual(transactionImportId)
			response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$.rows").Array().Length().IsEqual(2)

			// Rows must come back in the same order they appear in the CSV, so index
			// into them by position rather than searching by id. That way the test
			// also covers the row-order contract.
			rows := response.JSON().Path("$.rows").Array()
			coffee := rows.Value(0)
			gas := rows.Value(1)

			coffee.Path("$.data.id").String().IsEqual("COFFEE SHOP")
			coffee.Path("$.existingTransactionIds").Array().Length().IsEqual(1)
			coffee.Path("$.existingTransactionIds[0]").IsEqual(existingTransaction.TransactionId)

			// The GAS STATION row is the new one (no existing match), so the preview
			// should carry the parsed CSV data verbatim. monetr stores amounts in the
			// smallest currency unit, so -45.00 USD becomes -4500. Balance is 0
			// because the mapping uses kind=none, and posted is true by default when
			// the mapping has no posted spec.
			gas.Path("$.itemId").String().NotEmpty()
			gas.Path("$.existingTransactionIds").Array().Length().IsEqual(0)
			gas.Path("$.data.rowNumber").Number().IsEqual(2)
			gas.Path("$.data.id").String().IsEqual("GAS STATION")
			gas.Path("$.data.amount").Number().IsEqual(-4500)
			gas.Path("$.data.memo").String().IsEqual("GAS STATION")
			gas.Path("$.data.date").String().Contains("2026-01-16")
			gas.Path("$.data.posted").Boolean().IsTrue()
			gas.Path("$.data.balance").Number().IsEqual(0)
		}

		{ // The import itself should have moved to preview status.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionImportId", transactionImportId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.status").String().IsEqual("preview")
		}
	})
}
