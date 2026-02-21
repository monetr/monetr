package controller_test

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_lunch_flow"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestPostBankAccount(t *testing.T) {
	t.Run("create a bank account for a manual link", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId ID[Link]
		{ // Create the manual link
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "Manual Link",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkType").IsEqual(ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			response.JSON().Path("$.description").String().IsEqual("My personal link")
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
			assert.False(t, linkId.IsZero(), "must be able to extract the link ID")
		}

		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":           linkId,
					"availableBalance": 100,
					"currentBalance":   100,
					"limitBalance":     0,
					"mask":             "1234",
					"name":             "Checking Account",
					"originalName":     "PERSONAL CHECKING",
					"accountType":      DepositoryBankAccountType,
					"accountSubType":   CheckingBankAccountSubType,
					"status":           ActiveBankAccountStatus,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.availableBalance").Number().IsEqual(100)
			response.JSON().Path("$.currentBalance").Number().IsEqual(100)
			response.JSON().Path("$.limitBalance").Number().IsEqual(0)
			response.JSON().Path("$.mask").String().IsEqual("1234")
			response.JSON().Path("$.name").String().IsEqual("Checking Account")
			response.JSON().Path("$.originalName").String().IsEqual("PERSONAL CHECKING")
			response.JSON().Path("$.accountType").String().IsEqual(string(DepositoryBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CheckingBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(ActiveBankAccountStatus))
		}
	})

	t.Run("create a lunch flow bank account", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var lunchFlowLinkId ID[LunchFlowLink]
		{ // Create the lunch flow link first!
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "US Bank",
					"lunchFlowURL": "https://lunchflow.app/api/v1",
					"apiKey":       "foobar",
				}).
				Expect()

			response.Status(http.StatusOK)
			// Link should be in pending when its first created
			response.JSON().Path("$.status").IsEqual(LunchFlowLinkStatusPending)
			lunchFlowLinkId = ID[LunchFlowLink](response.JSON().Path("$.lunchFlowLinkId").String().Raw())
		}

		mock_lunch_flow.MockFetchAccounts(t, []lunch_flow.Account{
			{
				Id:              "1234",
				Name:            "Test Account",
				InstitutionName: "US Bank",
				Provider:        "Bogus",
				Currency:        "USD",
				Status:          "ACTIVE",
			},
		})

		mock_lunch_flow.MockFetchBalance(t, "1234", lunch_flow.Balance{
			Amount:   "1234.00",
			Currency: "USD",
		})

		{ // Refresh the accounts
			response := e.POST("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts/refresh").
				WithPath("lunchFlowLinkId", lunchFlowLinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNoContent)
			response.Body().IsEmpty()
		}

		var lunchFlowBankAccountId ID[LunchFlowBankAccount]
		{ // Check for bank account in the responsne
			response := e.GET("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts").
				WithPath("lunchFlowLinkId", lunchFlowLinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(1)
			lunchFlowBankAccountId = ID[LunchFlowBankAccount](response.JSON().Path("$[0].lunchFlowBankAccountId").String().Raw())
		}

		var linkId ID[Link]
		{ // Then create the actual link!
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
					"lunchFlowLinkId": lunchFlowLinkId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII()
			response.JSON().Path("$.linkType").IsEqual(models.LunchFlowLinkType)
			response.JSON().Path("$.lunchFlowLinkId").String().IsEqual(lunchFlowLinkId.String())
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
		}

		{ // Create the lunch flow bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":                 linkId,
					"lunchFlowBankAccountId": lunchFlowBankAccountId,
					"name":                   "Test account",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.lunchFlowBankAccountId").String().IsEqual(string(lunchFlowBankAccountId))
			response.JSON().Path("$.status").String().IsEqual(string(ActiveBankAccountStatus))
		}
	})

	t.Run("minimal creation", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId ID[Link]
		{ // Create the manual link
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "Manual Link",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkType").IsEqual(ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			response.JSON().Path("$.description").String().IsEqual("My personal link")
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
			assert.False(t, linkId.IsZero(), "must be able to extract the link ID")
		}

		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId": linkId,
					"name":   "Checking Account",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.availableBalance").Number().IsEqual(0)
			response.JSON().Path("$.currentBalance").Number().IsEqual(0)
			response.JSON().Path("$.limitBalance").Number().IsEqual(0)
			response.JSON().Path("$.mask").String().IsEmpty()
			response.JSON().Path("$.name").String().IsEqual("Checking Account")
			response.JSON().Path("$.originalName").String().IsEmpty()
			response.JSON().Path("$.accountType").String().IsEqual(string(DepositoryBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CheckingBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(ActiveBankAccountStatus))
		}
	})

	t.Run("create a bank account for a special locale", func(t *testing.T) {
		_, e := NewTestApplication(t)
		var token string
		{ // Register a new user
			email := testutils.GetUniqueEmail(t)
			password := gofakeit.Password(true, true, true, true, false, 32)
			response := e.POST(`/api/authentication/register`).
				WithJSON(map[string]any{
					"email":     email,
					"password":  password,
					"firstName": gofakeit.FirstName(),
					"lastName":  gofakeit.LastName(),
					// Create an account with a non-default locale such that the currency
					// code should be different.
					"locale":   "ja_JP",
					"timezone": "Asia/Tokyo",
				}).
				Expect()

			response.Status(http.StatusOK)
			token = GivenILogin(t, e, email, password)
		}

		var linkId ID[Link]
		{ // Create the manual link
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkType").IsEqual(ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			response.JSON().Path("$.description").String().IsEqual("My personal link")
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
			assert.False(t, linkId.IsZero(), "must be able to extract the link ID")
		}

		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":           linkId,
					"availableBalance": 100,
					"currentBalance":   100,
					"limitBalance":     0,
					"mask":             "1234",
					"name":             "Checking Account",
					"originalName":     "PERSONAL CHECKING",
					"accountType":      DepositoryBankAccountType,
					"accountSubType":   CheckingBankAccountSubType,
					"status":           ActiveBankAccountStatus,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.availableBalance").Number().IsEqual(100)
			response.JSON().Path("$.currentBalance").Number().IsEqual(100)
			response.JSON().Path("$.limitBalance").Number().IsEqual(0)
			response.JSON().Path("$.mask").String().IsEqual("1234")
			response.JSON().Path("$.name").String().IsEqual("Checking Account")
			response.JSON().Path("$.originalName").String().IsEqual("PERSONAL CHECKING")
			response.JSON().Path("$.accountType").String().IsEqual(string(DepositoryBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CheckingBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(ActiveBankAccountStatus))
			// Make sure that we do not default to USD when we have a locale with it's
			// own currency.
			response.JSON().Path("$.currency").String().IsEqual("JPY")
		}
	})

	t.Run("create a bank account overriding the currency code", func(t *testing.T) {
		_, e := NewTestApplication(t)
		var token string
		{ // Register a new user
			email := testutils.GetUniqueEmail(t)
			password := gofakeit.Password(true, true, true, true, false, 32)
			response := e.POST(`/api/authentication/register`).
				WithJSON(map[string]any{
					"email":     email,
					"password":  password,
					"firstName": gofakeit.FirstName(),
					"lastName":  gofakeit.LastName(),
					// Create an account with a non-default locale such that the currency
					// code should be different.
					"locale":   "ja_JP",
					"timezone": "Asia/Tokyo",
				}).
				Expect()

			response.Status(http.StatusOK)
			token = GivenILogin(t, e, email, password)
		}

		var linkId ID[Link]
		{ // Create the manual link
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkType").IsEqual(ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			response.JSON().Path("$.description").String().IsEqual("My personal link")
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
			assert.False(t, linkId.IsZero(), "must be able to extract the link ID")
		}

		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":           linkId,
					"availableBalance": 100,
					"currentBalance":   100,
					"limitBalance":     0,
					"mask":             "1234",
					"name":             "Checking Account",
					"originalName":     "PERSONAL CHECKING",
					"accountType":      DepositoryBankAccountType,
					"accountSubType":   CheckingBankAccountSubType,
					"status":           ActiveBankAccountStatus,
					"currency":         "EUR",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.availableBalance").Number().IsEqual(100)
			response.JSON().Path("$.currentBalance").Number().IsEqual(100)
			response.JSON().Path("$.limitBalance").Number().IsEqual(0)
			response.JSON().Path("$.mask").String().IsEqual("1234")
			response.JSON().Path("$.name").String().IsEqual("Checking Account")
			response.JSON().Path("$.originalName").String().IsEqual("PERSONAL CHECKING")
			response.JSON().Path("$.accountType").String().IsEqual(string(DepositoryBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CheckingBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(ActiveBankAccountStatus))
			// Because we specified a currency code in our request, and the currency
			// code is valid, we should respect it regardless of the client's actual
			// locale.
			response.JSON().Path("$.currency").String().IsEqual("EUR")
		}
	})

	t.Run("invalid currency code", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var linkId ID[Link]
		{ // Create the manual link
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "U.S. Bank",
					"description":     "My personal link",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkType").IsEqual(ManualLinkType)
			response.JSON().Path("$.institutionName").String().NotEmpty()
			response.JSON().Path("$.description").String().IsEqual("My personal link")
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
			assert.False(t, linkId.IsZero(), "must be able to extract the link ID")
		}

		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":           linkId,
					"availableBalance": 100,
					"currentBalance":   100,
					"limitBalance":     0,
					"mask":             "1234",
					"name":             "Checking Account",
					"originalName":     "PERSONAL CHECKING",
					"accountType":      DepositoryBankAccountType,
					"accountSubType":   CheckingBankAccountSubType,
					"status":           ActiveBankAccountStatus,
					"currency":         "???",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be one supported by the server")
		}
	})

	t.Run("requires a link", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": 100,
					"currentBalance":   100,
					"limitBalance":     0,
					"mask":             "1234",
					"name":             "Checking Account",
					"originalName":     "PERSONAL CHECKING",
					"accountType":      DepositoryBankAccountType,
					"accountSubType":   CheckingBankAccountSubType,
					"status":           ActiveBankAccountStatus,
				}).
				Expect()

			// This returns the same error as if you provide a valid link, because it
			// just sees that the link does not exist with a manual type. Not that the
			// link doesn't exist at all.
			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Invalid request")
			response.JSON().Path("$.problems.linkId").String().IsEqual("required key is missing")
		}
	})

	t.Run("requires a manual link", func(t *testing.T) {
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)
		plaidLink := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)

		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":           plaidLink.LinkId,
					"availableBalance": 100,
					"currentBalance":   100,
					"limitBalance":     0,
					"mask":             "1234",
					"name":             "Checking Account",
					"originalName":     "PERSONAL CHECKING",
					"accountType":      DepositoryBankAccountType,
					"accountSubType":   CheckingBankAccountSubType,
					"status":           ActiveBankAccountStatus,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Invalid request")
			response.JSON().Path("$.problems.linkId").String().IsEqual("Cannot create a bank account for a non-manual link, specify a manual Link ID")
		}
	})
}

func TestPatchBankAccount(t *testing.T) {
	t.Run("happy path patch a manual link bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": -100,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual(bank.Name)
			response.JSON().Path("$.currency").String().IsEqual(bank.Currency)
			response.JSON().Path("$.mask").String().IsEqual(bank.Mask)
			response.JSON().Path("$.availableBalance").Number().IsEqual(-100)
			response.JSON().Path("$.currentBalance").Number().IsEqual(bank.CurrentBalance)
			response.JSON().Path("$.status").String().IsEqual(string(bank.Status))
			response.JSON().Path("$.accountType").String().IsEqual(string(bank.AccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(bank.AccountSubType))
		}
	})
}

func TestPutBankAccount(t *testing.T) {
	t.Run("happy path update a manual link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PUT("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":     "My New Name",
					"currency": "USD",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual("My New Name")
			response.JSON().Path("$.currency").String().IsEqual("USD")
			response.JSON().Path("$.mask").String().IsEqual(bank.Mask)
			response.JSON().Path("$.availableBalance").Number().IsEqual(bank.AvailableBalance)
			response.JSON().Path("$.currentBalance").Number().IsEqual(bank.CurrentBalance)
			response.JSON().Path("$.status").String().IsEqual(string(bank.Status))
			response.JSON().Path("$.accountType").String().IsEqual(string(bank.AccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(bank.AccountSubType))
		}
	})

	t.Run("change currency", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // Make sure the bank account is created in USD.
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.currency").String().IsEqual("USD")
		}

		{ // Then update the bank account to be not USD.
			response := e.PUT("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":     "My New Name",
					"currency": "EUR",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual("My New Name")
			response.JSON().Path("$.currency").String().IsEqual("EUR")
			response.JSON().Path("$.mask").String().IsEqual(bank.Mask)
			response.JSON().Path("$.availableBalance").Number().IsEqual(bank.AvailableBalance)
			response.JSON().Path("$.currentBalance").Number().IsEqual(bank.CurrentBalance)
			response.JSON().Path("$.status").String().IsEqual(string(bank.Status))
			response.JSON().Path("$.accountType").String().IsEqual(string(bank.AccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(bank.AccountSubType))
		}
	})

	t.Run("invalid currency", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PUT("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": 1000,
					"currentBalance":   1000,
					"mask":             "1234",
					"name":             "My New Name",
					"currency":         "???",
					"status":           "active",
					"accountType":      "depository",
					"accountSubType":   "checking",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be one supported by the server")
		}
	})

	t.Run("invalid account type", func(t *testing.T) {
		t.Skip("not implemented yet")
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		fixtures.GivenIHaveNTransactions(t, app.Clock, bank, 10)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PUT("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": 1000,
					"currentBalance":   1000,
					"mask":             "1234",
					"name":             "My New Name",
					"currency":         "USD",
					"status":           "active",
					"accountType":      "something",
					"accountSubType":   "notchecking",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.accountType").String().IsEqual("Invalid bank account type")
			response.JSON().Path("$.problems.accountSubType").String().IsEqual("Invalid bank account sub type")
		}
	})

	t.Run("only update name for plaid bank accounts", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveAPlaidBankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PUT("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": 1000,
					"currentBalance":   1000,
					"mask":             "1234",
					"name":             "My New Name",
					"currency":         "USD",
					"status":           "active",
					"accountType":      "something",
					"accountSubType":   "notchecking",
				}).
				Expect()

			response.Status(http.StatusOK)
			// Non-updatable fields
			response.JSON().Path("$.availableBalance").Number().IsEqual(bank.AvailableBalance)
			response.JSON().Path("$.currentBalance").Number().IsEqual(bank.CurrentBalance)
			response.JSON().Path("$.mask").String().IsEqual(bank.Mask)
			response.JSON().Path("$.currency").String().IsEqual(bank.Currency)
			response.JSON().Path("$.status").String().IsEqual(string(bank.Status))
			response.JSON().Path("$.accountType").String().IsEqual(string(bank.AccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(bank.AccountSubType))

			response.JSON().Path("$.name").String().IsEqual("My New Name")
		}
	})
}

func TestDeleteBankAccount(t *testing.T) {
	t.Run("delete manual account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token = GivenILogin(t, e, user.Login.Email, password)

		{ // See the bank account in an API response
			response := e.GET("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Array().Length().IsEqual(1)
		}

		{ // Delete the bank account
			response := e.DELETE("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}

		{ // See the bank account in an API response
			response := e.GET("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}

		{ // Can still request by a single ID
			response := e.GET("/api/bank_accounts/{bankAccoundId}").
				WithPath("bankAccoundId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Path("$.deletedAt").String().NotEmpty()
		}
	})

	t.Run("cant delete Plaid account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveAPlaidBankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		token = GivenILogin(t, e, user.Login.Email, password)

		{ // See the bank account in an API response
			response := e.GET("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Array().Length().IsEqual(1)
		}

		{ // Delete the bank account
			response := e.DELETE("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Plaid bank account cannot be removed this way")
		}

		{ // See the bank account in an API response
			response := e.GET("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$[0].bankAccountId").IsEqual(bank.BankAccountId)
			response.JSON().Array().Length().IsEqual(1)
		}
	})

	t.Run("cant delete someone elses bank account", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		{ // Create a bank account under one user
			user, _ := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
			bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
		}

		{ // Create another user
			user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
			token = GivenILogin(t, e, user.Login.Email, password)
		}

		{ // See the bank account in an API response
			response := e.GET("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().IsEmpty()
		}

		{ // Delete the bank account
			response := e.DELETE("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").IsEqual("Failed to retrieve bank account: record does not exist")
		}
	})
}
