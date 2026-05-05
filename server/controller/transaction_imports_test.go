package controller_test

import (
	"net/http"
	"testing"

	"github.com/monetr/monetr/server/datasources/csv/csv_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostTransactionImport(t *testing.T) {
	t.Run("import CSV success", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		csvData := []byte(
			"Date,Description,Amount,Balance\n" +
				"2026-01-15,COFFEE SHOP,-4.50,1234.56\n" +
				"2026-01-16,GAS STATION,-45.00,1189.56\n",
		)

		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				testutils.NewGenericMatcher(func(file models.File) bool {
					return myownsanity.Every(
						assert.Equal(t, "transactions.csv", file.Name),
						assert.Equal(t, "transactions/imports", file.Kind),
						assert.Equal(t, bank.AccountId, file.AccountId),
						assert.Equal(t, models.TextCSVContentType, file.ContentType),
					)
				}),
			).
			Times(1).
			Return(nil)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.csv", csvData).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.bankAccountId").String().IsEqual(bank.BankAccountId.String())
		response.JSON().Path("$.fileId").String().NotEmpty()
		response.JSON().Path("$.delimeter").IsEqual(",")
		response.JSON().Path("$.headers").Array().IsEqual([]string{"Date", "Description", "Amount", "Balance"})
	})

	t.Run("rejects non-CSV input", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "not-a-csv.txt", []byte("this is just a sentence with no delimiters whatsoever")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Failed to parse CSV file")
	})

	t.Run("rejects multiple files in the same request", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "first.csv", []byte("Date,Description,Amount\n2026-01-15,COFFEE,-4.50\n")).
			WithFileBytes("data", "second.csv", []byte("Date,Description,Amount\n2026-01-16,GAS,-45.00\n")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("exactly one file must be uploaded under field \"data\"")
	})

	t.Run("storage disabled", func(t *testing.T) {
		config := NewTestApplicationConfig(t)
		config.Storage.Enabled = false
		app, e := NewTestApplicationWithConfig(t, config)
		var token string
		var bank BankAccount

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.csv", []byte("Date,Description,Amount\n2026-01-15,COFFEE,-4.50\n")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("File uploads are not enabled on this server")
	})

	t.Run("requires authentication", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
			WithPath("bankAccountId", "bac_unauthorized_test").
			WithMultipart().
			WithFileBytes("data", "transactions.csv", []byte("Date,Description,Amount\n2026-01-15,COFFEE,-4.50\n")).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestGetTransactionImport(t *testing.T) {
	t.Run("retrieve a transaction import by ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(1).
			Return(nil)

		var transactionImportId ID[TransactionImport]
		{ // Create the transaction import.
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
				WithPath("bankAccountId", bank.BankAccountId).
				WithMultipart().
				WithFileBytes("data", "transactions.csv", []byte("Date,Description,Amount\n2026-01-15,COFFEE,-4.50\n")).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.transactionImportId").String().IsASCII()

			// Save the ID of the created transaction import so we can use it below.
			transactionImportId = ID[TransactionImport](response.JSON().Path("$.transactionImportId").String().Raw())
		}

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionImportId", transactionImportId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transactionImportId").IsEqual(transactionImportId)
		response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
		response.JSON().Path("$.fileId").String().NotEmpty()
		response.JSON().Path("$.headers").Array().IsEqual([]string{"Date", "Description", "Amount"})
		response.JSON().Path("$.status").IsEqual("mapping")
	})

	t.Run("invalid bank account ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
			WithPath("bankAccountId", "bogus_bank_id").
			WithPath("transactionImportId", "txim_anything").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
	})

	t.Run("invalid transaction import ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionImportId", "bogus_import").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid transaction import Id")
	})

	t.Run("zero transaction import ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		// A bare prefix parses without error but is zero, so it must be
		// rejected before it reaches the database.
		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionImportId", "txim_").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid transaction import Id")
	})

	t.Run("non-existant transaction import", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionImportId", "txim_bogus").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to retrieve transaction import: record does not exist")
	})

	t.Run("cant get a transaction import from a different bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bankOne := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		bankTwo := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token := GivenILogin(t, e, user.Login.Email, password)

		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(1).
			Return(nil)

		var transactionImportId ID[TransactionImport]
		{ // Create the transaction import under the first bank account.
			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
				WithPath("bankAccountId", bankOne.BankAccountId).
				WithMultipart().
				WithFileBytes("data", "transactions.csv", []byte("Date,Description,Amount\n2026-01-15,COFFEE,-4.50\n")).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			transactionImportId = ID[TransactionImport](response.JSON().Path("$.transactionImportId").String().Raw())
		}

		{ // Then try to read it under the second bank account.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
				WithPath("bankAccountId", bankTwo.BankAccountId).
				WithPath("transactionImportId", transactionImportId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve transaction import: record does not exist")
		}
	})

	t.Run("cant get someone elses transaction import by ID", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount
		var transactionImportId ID[TransactionImport]

		{ // Create a bank account and transaction import under one user.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			tok := GivenILogin(t, e, user.Login.Email, password)

			app.Storage.EXPECT().
				Store(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).
				Times(1).
				Return(nil)

			response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/import").
				WithPath("bankAccountId", bank.BankAccountId).
				WithMultipart().
				WithFileBytes("data", "transactions.csv", []byte("Date,Description,Amount\n2026-01-15,COFFEE,-4.50\n")).
				WithCookie(TestCookieName, tok).
				Expect()

			response.Status(http.StatusOK)
			transactionImportId = ID[TransactionImport](response.JSON().Path("$.transactionImportId").String().Raw())
		}

		{ // Create another user.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // Try to get the transaction import using the other user's bank account and import IDs.
			response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionImportId", transactionImportId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").String().IsEqual("failed to retrieve transaction import: record does not exist")
		}
	})
}

func TestPatchTransactionImport(t *testing.T) {
	t.Run("selecting a mapping kicks off the preview job", func(t *testing.T) {
		// End-to-end happy path for the mapping -> pending-preview transition. The
		// user uploads a CSV (which lands in the mapping status), creates a
		// TransactionImportMapping that lines up with the file's headers, then
		// PATCHes the import with that mapping id and status=pending-preview. The
		// controller must persist the new status and enqueue
		// [csv_jobs.PreviewCSVImport] inside the same transaction it used for the
		// patch, with the right Account/BankAccount/TransactionImport ids in the
		// args.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		csvData := []byte(
			"Date,Description,Amount\n" +
				"2026-01-15,COFFEE SHOP,-4.50\n" +
				"2026-01-16,GAS STATION,-45.00\n",
		)

		// The POST upload triggers exactly one Store call. We don't care about the
		// file contents at this layer, just that the upload path runs.
		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(1).
			Return(nil)

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
			transactionImportId = ID[TransactionImport](response.JSON().Path("$.transactionImportId").String().Raw())
		}

		var mappingId ID[TransactionImportMapping]
		{ // Create a mapping whose headers line up with the uploaded CSV. The id is
			// hashed across all three columns since the file does not carry its own
			// stable identifier.
			response := e.POST("/api/mappings").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"mapping": map[string]any{
						"id": map[string]any{
							"kind": "hashed",
							"fields": []any{
								map[string]any{
									"name": "Date",
								},
								map[string]any{
									"name": "Description",
								},
								map[string]any{
									"name": "Amount",
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
		// the EnqueueAt call lands on the transactional queue. Returning app.Queue
		// from WithTransaction is the convention the OFX upload tests use too.
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
			Return(nil)

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
	})

	t.Run("invalid status transition", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		csvData := []byte(
			"Date,Description,Amount\n" +
				"2026-01-15,COFFEE SHOP,-4.50\n" +
				"2026-01-16,GAS STATION,-45.00\n",
		)

		// The POST upload triggers exactly one Store call. We don't care about the
		// file contents at this layer, just that the upload path runs.
		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(1).
			Return(nil)

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
			transactionImportId = ID[TransactionImport](response.JSON().Path("$.transactionImportId").String().Raw())
		}

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
			Times(0).
			Return(nil)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionImportId", transactionImportId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"status": TransactionImportStatusPendingProcessing,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("Cannot move a transaction import to any status other than pending preview from mapping")
	})

	t.Run("no cross account patching", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		csvData := []byte(
			"Date,Description,Amount\n" +
				"2026-01-15,COFFEE SHOP,-4.50\n" +
				"2026-01-16,GAS STATION,-45.00\n",
		)

		// The POST upload triggers exactly one Store call. We don't care about the
		// file contents at this layer, just that the upload path runs.
		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(1).
			Return(nil)

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
			transactionImportId = ID[TransactionImport](response.JSON().Path("$.transactionImportId").String().Raw())
		}

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
			Times(0).
			Return(nil)

		{
			// Create another use and try to patch the original import
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)

			// Create the mapping under the new user
			var mappingId ID[TransactionImportMapping]
			{
				response := e.POST("/api/mappings").
					WithCookie(TestCookieName, token).
					WithJSON(map[string]any{
						"mapping": map[string]any{
							"id": map[string]any{
								"kind": "hashed",
								"fields": []any{
									map[string]any{
										"name": "Date",
									},
									map[string]any{
										"name": "Description",
									},
									map[string]any{
										"name": "Amount",
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

			response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithPath("transactionImportId", transactionImportId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"transactionImportMappingId": mappingId,
					"status":                     TransactionImportStatusPendingPreview,
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").IsEqual("failed to retrieve transaction import: record does not exist")
		}
	})

	t.Run("patch invalid field", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed data
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		csvData := []byte(
			"Date,Description,Amount\n" +
				"2026-01-15,COFFEE SHOP,-4.50\n" +
				"2026-01-16,GAS STATION,-45.00\n",
		)

		// The POST upload triggers exactly one Store call. We don't care about the
		// file contents at this layer, just that the upload path runs.
		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				gomock.Any(),
			).
			Times(1).
			Return(nil)

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
			transactionImportId = ID[TransactionImport](response.JSON().Path("$.transactionImportId").String().Raw())
		}

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
			Times(0).
			Return(nil)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}/transactions/import/{transactionImportId}").
			WithPath("bankAccountId", bank.BankAccountId).
			WithPath("transactionImportId", transactionImportId).
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"fileId": "file_bogus",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("Invalid request")
	})
}
