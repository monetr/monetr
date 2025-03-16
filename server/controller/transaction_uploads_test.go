package controller_test

import (
	"net/http"
	"testing"

	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	. "github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostTransaactionUpload(t *testing.T) {
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
				gomock.Eq(storage.FileInfo{
					Name:        "transactions.ofx",
					Kind:        "transactions/uploads",
					AccountId:   bank.AccountId,
					ContentType: storage.IntuitQFXContentType,
				}),
			).
			Times(1).
			Return(
				"blob:///bogus.ofx",
				nil,
			)

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				background.ProcessOFXUpload,
				testutils.NewGenericMatcher(func(args background.ProcessOFXUploadArguments) bool {
					a := assert.EqualValues(t, bank.AccountId, args.AccountId, "Account ID should match")
					b := assert.EqualValues(t, bank.BankAccountId, args.BankAccountId, "Bank Account ID should match")
					return a && b
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
				"",
				nil,
			)

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				background.ProcessOFXUpload,
				gomock.Any(),
			).
			Times(0).
			Return(nil)

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
				"blob:///bogus.ofx",
				nil,
			)

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				background.ProcessOFXUpload,
				gomock.Any(),
			).
			Times(0).
			Return(nil)

		response := e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("File uploads are not enabled on this server")
	})
}
