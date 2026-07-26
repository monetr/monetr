package controller_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/datasources/ofx/ofx_jobs"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockqueue"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostTransactionUpload(t *testing.T) {
	t.Run("upload OFX file success", func(t *testing.T) {
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
				testutils.NewGenericMatcher(func(file File) bool {
					return myownsanity.Every(
						assert.Equal(t, "transactions.ofx", file.Name),
						assert.Equal(t, "transactions/uploads", file.Kind),
						assert.Equal(t, bank.AccountId, file.AccountId),
						assert.Equal(t, IntuitQFXContentType, file.ContentType),
					)
				}),
			).
			Times(1).
			Return(nil)

		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(ofx_jobs.ProcessOFXUpload),
				gomock.Any(),
				testutils.NewGenericMatcher(func(args ofx_jobs.ProcessOFXUploadArguments) bool {
					return myownsanity.Every(
						assert.EqualValues(t, bank.AccountId, args.AccountId, "Account ID should match"),
						assert.EqualValues(t, bank.BankAccountId, args.BankAccountId, "Bank Account ID should match"),
					)
				}),
			).
			Times(1).
			Return(nil)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transactionUploadId").String().NotEmpty()
		response.JSON().Path("$.bankAccountId").String().IsEqual(bank.BankAccountId.String())
		response.JSON().Path("$.fileId").String().NotEmpty()
		response.JSON().Path("$.status").String().IsEqual("pending")
	})

	t.Run("fails to upload a malformed body", func(t *testing.T) {
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

		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(ofx_jobs.ProcessOFXUpload),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("bogus", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Failed to read file upload")
	})

	t.Run("upload camt.053 to the wrong endpoint", func(t *testing.T) {
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
			Times(0).
			Return(
				nil,
			)

		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(ofx_jobs.ProcessOFXUpload),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "statement.xml", fixtures.LoadFile(t, "goldman-us-camt053-v2.xml")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").IsEqual("Unsupported file type!")
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
			Times(0).
			Return(
				nil,
			)

		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(ofx_jobs.ProcessOFXUpload),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("File uploads are not enabled on this server")
	})

	t.Run("storage failure", func(t *testing.T) {
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
				testutils.NewGenericMatcher(func(file File) bool {
					return myownsanity.Every(
						assert.Equal(t, "transactions.ofx", file.Name),
						assert.Equal(t, "transactions/uploads", file.Kind),
						assert.Equal(t, bank.AccountId, file.AccountId),
						assert.Equal(t, IntuitQFXContentType, file.ContentType),
					)
				}),
			).
			Times(1).
			Return(
				errors.New("no space available"),
			)

		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(ofx_jobs.ProcessOFXUpload),
				gomock.Any(),
				gomock.Any(),
			).
			Times(0)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().IsEqual("Failed to upload file")
	})

	t.Run("with a valid api key", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Seed data, cloning the OFX upload happy path.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		// The API key belongs to the same account that seeded the data.
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		app.Storage.EXPECT().
			Store(
				gomock.Any(),
				gomock.Any(),
				testutils.NewGenericMatcher(func(file File) bool {
					return myownsanity.Every(
						assert.Equal(t, "transactions.ofx", file.Name),
						assert.Equal(t, "transactions/uploads", file.Kind),
						assert.Equal(t, bank.AccountId, file.AccountId),
						assert.Equal(t, IntuitQFXContentType, file.ContentType),
					)
				}),
			).
			Times(1).
			Return(nil)

		app.Queue.EXPECT().
			WithTransaction(
				gomock.Any(),
			).
			Return(app.Queue)
		app.Queue.EXPECT().
			EnqueueAt(
				gomock.Any(),
				mockqueue.EqQueue(ofx_jobs.ProcessOFXUpload),
				gomock.Any(),
				testutils.NewGenericMatcher(func(args ofx_jobs.ProcessOFXUploadArguments) bool {
					return myownsanity.Every(
						assert.EqualValues(t, bank.AccountId, args.AccountId, "Account ID should match"),
						assert.EqualValues(t, bank.BankAccountId, args.BankAccountId, "Bank Account ID should match"),
					)
				}),
			).
			Times(1).
			Return(nil)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transactionUploadId").String().NotEmpty()
		response.JSON().Path("$.bankAccountId").String().IsEqual(bank.BankAccountId.String())
		response.JSON().Path("$.fileId").String().NotEmpty()
		response.JSON().Path("$.status").String().IsEqual("pending")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		// The authentication middleware rejects the unknown key before the handler
		// runs, so no file is ever stored and no job is ever enqueued.
		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", "bac_fake").
			WithMultipart().
			WithFileBytes("data", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestGetTransactionUpload(t *testing.T) {
	t.Run("with a valid api key", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var upload TransactionUpload

		{ // Seed data. Insert a file and a completed transaction upload directly so
			// there is something to read back by its Id.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

			file := testutils.MustInsert(t, File{
				AccountId:   user.AccountId,
				Kind:        TransactionUpload{}.FileKind(),
				Name:        "transactions.ofx",
				ContentType: IntuitQFXContentType,
				Size:        1234,
				CreatedAt:   app.Clock.Now(),
				CreatedBy:   user.UserId,
			})
			upload = testutils.MustInsert(t, TransactionUpload{
				AccountId:     user.AccountId,
				BankAccountId: bank.BankAccountId,
				FileId:        file.FileId,
				Status:        TransactionUploadStatusComplete,
				CreatedAt:     app.Clock.Now(),
				CreatedBy:     user.UserId,
			})

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/upload/{transactionUploadId}").
			WithPath("bankAccountId", upload.BankAccountId).
			WithPath("transactionUploadId", upload.TransactionUploadId).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.transactionUploadId").String().IsEqual(upload.TransactionUploadId.String())
		response.JSON().Path("$.status").String().IsEqual("complete")
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/upload/{transactionUploadId}").
			WithPath("bankAccountId", "bac_fake").
			WithPath("transactionUploadId", "txup_fake").
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}

func TestGetTransactionUploadProgress(t *testing.T) {
	// The progress endpoint upgrades the request to a websocket connection. A
	// successful upgrade responds with 101 Switching Protocols rather than a 2xx
	// status, so the valid case asserts the handshake succeeds and a message is
	// received. This still proves the API key is accepted, because an invalid key
	// is rejected with a 401 by the authentication middleware before the handshake
	// ever happens.
	t.Run("with a valid api key", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var upload TransactionUpload

		{ // Seed data. A completed upload makes the websocket send its final status
			// and close promptly instead of waiting for progress notifications.
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

			file := testutils.MustInsert(t, File{
				AccountId:   user.AccountId,
				Kind:        TransactionUpload{}.FileKind(),
				Name:        "transactions.ofx",
				ContentType: IntuitQFXContentType,
				Size:        1234,
				CreatedAt:   app.Clock.Now(),
				CreatedBy:   user.UserId,
			})
			upload = testutils.MustInsert(t, TransactionUpload{
				AccountId:     user.AccountId,
				BankAccountId: bank.BankAccountId,
				FileId:        file.FileId,
				Status:        TransactionUploadStatusComplete,
				CreatedAt:     app.Clock.Now(),
				CreatedBy:     user.UserId,
			})

			token = GivenILogin(t, e, user.Login.Email, password)
		}

		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)

		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/upload/{transactionUploadId}/progress").
			WithPath("bankAccountId", upload.BankAccountId).
			WithPath("transactionUploadId", upload.TransactionUploadId).
			WithBasicAuth(apiKeyId, apiKeySecret).
			// The server's websocket handshake requires an Origin header, the default
			// websocket client does not set one.
			WithHeader("Origin", "https://monetr.local").
			WithWebsocketUpgrade().
			Expect()

		// A valid API key is accepted and the websocket handshake completes.
		response.Status(http.StatusSwitchingProtocols)
		ws := response.Websocket()
		defer ws.Disconnect()

		// The server pushes the current upload state as the first message, proving
		// the handler ran to completion under API key authentication.
		ws.Expect().TextMessage()
		ws.Close()
	})

	t.Run("with an invalid api key", func(t *testing.T) {
		_, e := NewTestApplication(t)

		// A plain request with an unknown key is rejected with a 401 before any
		// websocket upgrade is attempted.
		response := e.GET("/api/bank_accounts/{bankAccountId}/transactions/upload/{transactionUploadId}/progress").
			WithPath("bankAccountId", "bac_fake").
			WithPath("transactionUploadId", "txup_fake").
			WithBasicAuth("key_"+gofakeit.UUID(), "monetr_secret_"+gofakeit.UUID()).
			Expect()

		response.Status(http.StatusUnauthorized)
	})
}
