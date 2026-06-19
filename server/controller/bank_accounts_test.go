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
					"status":           BankAccountStatusActive,
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
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
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
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
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

		var bankAccountId ID[BankAccount]
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
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
			bankAccountId = ID[BankAccount](response.JSON().Path("$.bankAccountId").String().Raw())
		}

		{ // Read the bank account back to make sure its in the right status
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.lunchFlowBankAccountId").String().IsEqual(string(lunchFlowBankAccountId))
			response.JSON().Path("$.lunchFlowBankAccount.lunchFlowBankAccountId").String().IsEqual(string(lunchFlowBankAccountId))
			response.JSON().Path("$.lunchFlowBankAccount.status").String().IsEqual(string(LunchFlowBankAccountStatusActive))
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
		}
	})

	t.Run("cannot create a lunch flow bank account from another lunch flow link", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		var firstLunchFlowLinkId ID[LunchFlowLink]
		{ // Create the first lunch flow link.
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "First Lunch Flow Account",
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
					"apiKey":       "firstapikey",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.status").IsEqual(LunchFlowLinkStatusPending)
			firstLunchFlowLinkId = ID[LunchFlowLink](response.JSON().Path("$.lunchFlowLinkId").String().Raw())
		}

		mock_lunch_flow.MockFetchAccounts(t, []lunch_flow.Account{
			{
				Id:              "1234",
				Name:            "First Lunch Flow Account",
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

		{ // Refresh the first lunch flow link accounts.
			response := e.POST("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts/refresh").
				WithPath("lunchFlowLinkId", firstLunchFlowLinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNoContent)
			response.Body().IsEmpty()
		}

		var firstLunchFlowBankAccountId ID[LunchFlowBankAccount]
		{ // Read the first lunch flow bank account.
			response := e.GET("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts").
				WithPath("lunchFlowLinkId", firstLunchFlowLinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(1)
			firstLunchFlowBankAccountId = ID[LunchFlowBankAccount](response.JSON().Path("$[0].lunchFlowBankAccountId").String().Raw())
		}

		var firstLinkId ID[Link]
		{ // Create the first actual link.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "First Lunch Flow Account",
					"description":     "My personal link",
					"lunchFlowLinkId": firstLunchFlowLinkId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkType").IsEqual(models.LunchFlowLinkType)
			response.JSON().Path("$.lunchFlowLinkId").String().IsEqual(firstLunchFlowLinkId.String())
			firstLinkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
		}

		var secondLunchFlowLinkId ID[LunchFlowLink]
		{ // Create the second lunch flow link.
			response := e.POST("/api/lunch_flow/link").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":         "Second Lunch Flow Account",
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
					"apiKey":       "secondapikey",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.status").IsEqual(LunchFlowLinkStatusPending)
			secondLunchFlowLinkId = ID[LunchFlowLink](response.JSON().Path("$.lunchFlowLinkId").String().Raw())
		}

		mock_lunch_flow.MockFetchAccounts(t, []lunch_flow.Account{
			{
				Id:              "9876",
				Name:            "Second Lunch Flow Account",
				InstitutionName: "US Bank",
				Provider:        "Bogus",
				Currency:        "USD",
				Status:          "ACTIVE",
			},
		})
		mock_lunch_flow.MockFetchBalance(t, "9876", lunch_flow.Balance{
			Amount:   "9876.00",
			Currency: "USD",
		})

		{ // Refresh the second lunch flow link accounts.
			response := e.POST("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts/refresh").
				WithPath("lunchFlowLinkId", secondLunchFlowLinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusNoContent)
			response.Body().IsEmpty()
		}

		var secondLunchFlowBankAccountId ID[LunchFlowBankAccount]
		{ // Read the second lunch flow bank account.
			response := e.GET("/api/lunch_flow/link/{lunchFlowLinkId}/bank_accounts").
				WithPath("lunchFlowLinkId", secondLunchFlowLinkId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Array().Length().IsEqual(1)
			secondLunchFlowBankAccountId = ID[LunchFlowBankAccount](response.JSON().Path("$[0].lunchFlowBankAccountId").String().Raw())
		}

		var secondLinkId ID[Link]
		{ // Create the second actual link.
			response := e.POST("/api/links").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"institutionName": "Second Lunch Flow Account",
					"description":     "My personal link",
					"lunchFlowLinkId": secondLunchFlowLinkId,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkType").IsEqual(models.LunchFlowLinkType)
			response.JSON().Path("$.lunchFlowLinkId").String().IsEqual(secondLunchFlowLinkId.String())
			secondLinkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
		}

		{ // Try to create a bank account with the first link and a Lunch Flow bank account from the second link.
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":                 firstLinkId,
					"lunchFlowBankAccountId": secondLunchFlowBankAccountId,
					"name":                   "Wrong account",
				}).
				Expect()

			// This check moved out of the schema and into the handler, so it now comes
			// back as a plain bad request instead of a per field problem.
			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Lunch Flow Bank Account ID must belong to the specified link")
			response.JSON().Object().NotContainsKey("problems")
		}

		{ // Make sure the first link still accepts the first Lunch Flow bank account.
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":                 firstLinkId,
					"lunchFlowBankAccountId": firstLunchFlowBankAccountId,
					"name":                   "First account",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsEqual(firstLinkId.String())
			response.JSON().Path("$.lunchFlowBankAccountId").String().IsEqual(string(firstLunchFlowBankAccountId))
		}

		{ // Make sure the second link still accepts the second Lunch Flow bank account.
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":                 secondLinkId,
					"lunchFlowBankAccountId": secondLunchFlowBankAccountId,
					"name":                   "Second account",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.linkId").String().IsEqual(secondLinkId.String())
			response.JSON().Path("$.lunchFlowBankAccountId").String().IsEqual(string(secondLunchFlowBankAccountId))
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
			response.JSON().Path("$.mask").IsNull()
			response.JSON().Path("$.name").String().IsEqual("Checking Account")
			response.JSON().Path("$.originalName").String().IsEmpty()
			response.JSON().Path("$.accountType").String().IsEqual(string(DepositoryBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CheckingBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
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
					"status":           BankAccountStatusActive,
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
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
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
					"status":           BankAccountStatusActive,
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
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
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
					"status":           BankAccountStatusActive,
					"currency":         "???",
				}).
				Expect()

			// "???" is three characters so it gets past the length rule, but it trips
			// the alphabetical rule that the new currency schema added.
			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be alphabetical characters only")
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
					"status":           BankAccountStatusActive,
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
					"status":           BankAccountStatusActive,
				}).
				Expect()

			// The non-manual link rejection happens in the handler now, so it is a
			// plain bad request rather than a per field problem.
			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Cannot create a bank account for a non-manual link, specify a manual Link ID")
			response.JSON().Object().NotContainsKey("problems")
		}
	})

	t.Run("invalid link Id format", func(t *testing.T) {
		// The new schema validates the shape of the link Id before we ever go to
		// the database. A value that is not even shaped like one of our IDs should
		// come back as a problem on the linkId field.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/bank_accounts").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId": "potato",
				"name":   "Checking Account",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.linkId").String().IsEqual("id does not match format link_...")
	})

	t.Run("mask must be exactly four digits", func(t *testing.T) {
		// The old schema only checked that the mask contained four digits
		// somewhere, so a longer string would slip through. The new schema pins it
		// to exactly four digits. The linkId here is a valid shape but never gets
		// looked up because validation fails first.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/bank_accounts").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId": "link_01hy4rbb1gjdek7h2xmgy5pnwk",
				"name":   "Checking Account",
				"mask":   "12345",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		// Mask is a one of null or a real mask, and the one of error now serializes as
		// a structured envelope with a branch per alternative. We only care that the
		// four digit rule shows up as the reason the real mask branch failed.
		response.JSON().Path("$.problems.mask.oneOf").Array().ContainsAll("Mask must be exactly 4 digits")
	})

	t.Run("mask can be explicitly null", func(t *testing.T) {
		// The mask field is now a one of null or a four digit string, so a client
		// that sends an explicit null should be totally fine. This used to not have
		// an explicit null branch.
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
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
			assert.False(t, linkId.IsZero(), "must be able to extract the link ID")
		}

		{ // Create the bank account with an explicit null mask
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId": linkId,
					"name":   "Checking Account",
					"mask":   nil,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.mask").IsNull()
		}
	})

	t.Run("currency must be all upper case", func(t *testing.T) {
		// The new currency schema is much stricter than the old one which only
		// checked that the value was in the installed list. A lower case currency
		// now trips the upper case rule with its own specific message.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/bank_accounts").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId":   "link_01hy4rbb1gjdek7h2xmgy5pnwk",
				"name":     "Checking Account",
				"currency": "usd",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be all upper case")
	})

	t.Run("currency must be exactly three characters", func(t *testing.T) {
		// Same idea, a currency that is the wrong length gets its own message now
		// instead of just being reported as unsupported.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/bank_accounts").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId":   "link_01hy4rbb1gjdek7h2xmgy5pnwk",
				"name":     "Checking Account",
				"currency": "US",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be exactly 3 characters long")
	})

	t.Run("limit balance cannot be negative", func(t *testing.T) {
		// Limit balance is now a one of null or a non-negative integer. A negative
		// value should be rejected.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/bank_accounts").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId":       "link_01hy4rbb1gjdek7h2xmgy5pnwk",
				"name":         "Checking Account",
				"limitBalance": -5,
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		// Same structured one of envelope as the mask, we just want the negative branch
		// message to be present.
		response.JSON().Path("$.problems.limitBalance.oneOf").Array().ContainsAll("Limit balance cannot be negative")
	})

	t.Run("balance must be an integer", func(t *testing.T) {
		// Current and available balance only need to be integers, but they DO need
		// to be integers. A string value should be rejected as a non integer.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)
		response := e.POST("/api/bank_accounts").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"linkId":         "link_01hy4rbb1gjdek7h2xmgy5pnwk",
				"name":           "Checking Account",
				"currentBalance": "not a number",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("Invalid request")
		response.JSON().Path("$.problems.currentBalance").String().IsEqual("must be an integer")
	})

	t.Run("merges non-default fields over the controller defaults", func(t *testing.T) {
		// The schema only validates a subset of the fields, but the merge step
		// applies the WHOLE request body onto the bank account. The controller
		// pre-sets some defaults (depository, checking, active) before parsing, so
		// this test sends non-default values for all of those to prove the merge
		// actually overrides the defaults rather than the defaults winning. We read
		// the account back afterwards to make sure the merged values were persisted
		// and not just echoed.
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
			linkId = ID[Link](response.JSON().Path("$.linkId").String().Raw())
			assert.False(t, linkId.IsZero(), "must be able to extract the link ID")
		}

		var bankAccountId ID[BankAccount]
		{ // Create the bank account with everything set to a non-default value
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"linkId":           linkId,
					"name":             "Credit Card Account",
					"originalName":     "ORIGINAL CREDIT",
					"mask":             "9876",
					"availableBalance": 1500,
					"currentBalance":   1200,
					"limitBalance":     5000,
					"accountType":      CreditBankAccountType,
					"accountSubType":   CreditCardBankAccountSubType,
					"status":           BankAccountStatusInactive,
				}).
				Expect()

			response.Status(http.StatusOK)
			bankAccountId = ID[BankAccount](response.JSON().Path("$.bankAccountId").String().Raw())
			assert.False(t, bankAccountId.IsZero(), "must be able to extract the bank account ID")
			// Every one of these is a non-default value, so seeing them come back
			// proves the merge applied the request body over the pre-set defaults.
			response.JSON().Path("$.originalName").String().IsEqual("ORIGINAL CREDIT")
			response.JSON().Path("$.mask").String().IsEqual("9876")
			response.JSON().Path("$.availableBalance").Number().IsEqual(1500)
			response.JSON().Path("$.currentBalance").Number().IsEqual(1200)
			response.JSON().Path("$.limitBalance").Number().IsEqual(5000)
			response.JSON().Path("$.accountType").String().IsEqual(string(CreditBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CreditCardBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusInactive))
		}

		{ // Read it back to make sure the merged values were actually persisted
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.originalName").String().IsEqual("ORIGINAL CREDIT")
			response.JSON().Path("$.mask").String().IsEqual("9876")
			response.JSON().Path("$.availableBalance").Number().IsEqual(1500)
			response.JSON().Path("$.currentBalance").Number().IsEqual(1200)
			response.JSON().Path("$.limitBalance").Number().IsEqual(5000)
			response.JSON().Path("$.accountType").String().IsEqual(string(CreditBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CreditCardBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusInactive))
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

		{ // A manual link bank account has user managed balances and
			// classification, so it is allowed to change its name, mask, currency,
			// balances, and account type/sub type.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":             "My New Name",
					"mask":             "4321",
					"currency":         "EUR",
					"availableBalance": -100,
					"currentBalance":   200,
					"limitBalance":     5000,
					"accountType":      CreditBankAccountType,
					"accountSubType":   CreditCardBankAccountSubType,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual("My New Name")
			response.JSON().Path("$.mask").String().IsEqual("4321")
			response.JSON().Path("$.currency").String().IsEqual("EUR")
			response.JSON().Path("$.availableBalance").Number().IsEqual(-100)
			response.JSON().Path("$.currentBalance").Number().IsEqual(200)
			response.JSON().Path("$.limitBalance").Number().IsEqual(5000)
			response.JSON().Path("$.accountType").String().IsEqual(string(CreditBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CreditCardBankAccountSubType))
			// Status is intentionally not part of the manual patch schema so it
			// should come back exactly as it was seeded.
			response.JSON().Path("$.status").String().IsEqual(string(bank.Status))
		}

		{ // Read it back to make sure the patched values were actually persisted
			// and not just echoed back from the handler.
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual("My New Name")
			response.JSON().Path("$.mask").String().IsEqual("4321")
			response.JSON().Path("$.currency").String().IsEqual("EUR")
			response.JSON().Path("$.availableBalance").Number().IsEqual(-100)
			response.JSON().Path("$.currentBalance").Number().IsEqual(200)
			response.JSON().Path("$.limitBalance").Number().IsEqual(5000)
			response.JSON().Path("$.accountType").String().IsEqual(string(CreditBankAccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(CreditCardBankAccountSubType))
			response.JSON().Path("$.status").String().IsEqual(string(bank.Status))
		}
	})

	t.Run("manual bank account cannot patch the status", func(t *testing.T) {
		// The manual patch schema deliberately leaves status out, monetr owns it.
		// So even though balances and the account type can be changed, status is a
		// key the schema does not expect and the request should be rejected before
		// we ever touch the database.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // Try to patch the status.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"status": string(BankAccountStatusInactive),
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.status").String().IsEqual("key not expected")
		}
	})

	t.Run("manual bank account rejects a negative limit balance", func(t *testing.T) {
		// The limit balance is a one of null or a non negative integer just like on
		// the create path, so a negative value should be rejected.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"limitBalance": -5,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.limitBalance.oneOf").Array().ContainsAll("Limit balance cannot be negative")
		}
	})

	t.Run("manual bank account rejects an invalid account type", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"accountType": "something",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.accountType").String().IsEqual("Invalid bank account type")
		}
	})

	t.Run("manual bank account can clear the mask", func(t *testing.T) {
		// The mask field on the manual patch schema is a one of null or a real
		// mask, so a client should be able to explicitly null it out and actually
		// have it cleared. This used to silently leave the existing mask in place
		// for two reasons that have both since been fixed. The merge code skipped
		// nil values, and the repository used UpdateNotZero which would not write a
		// nil column back as NULL.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // Make sure the mask starts off set.
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.mask").String().IsEqual(*bank.Mask)
		}

		{ // Then null it out.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"mask": nil,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.mask").IsNull()
		}

		{ // And make sure it actually persisted as null.
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.mask").IsNull()
		}
	})

	t.Run("manual bank account can set a balance back to zero", func(t *testing.T) {
		// Balances are non-nullable integers where zero is a totally valid value.
		// We set a balance to a non-zero number and then back to zero to prove the
		// zero actually persists. This is the case the old UpdateNotZero behavior
		// got wrong, it would have skipped writing the zero and left the previous
		// value in place. The full Update fixes that.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // First set the limit balance to a non-zero value.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"limitBalance": 5000,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.limitBalance").Number().IsEqual(5000)
		}

		{ // Then set it back to zero.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"limitBalance": 0,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.limitBalance").Number().IsEqual(0)
		}

		{ // And make sure the zero actually persisted rather than being skipped.
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.limitBalance").Number().IsEqual(0)
		}
	})

	t.Run("empty patch is a no-op", func(t *testing.T) {
		// An empty patch body should be accepted and change nothing. Every field is
		// optional so there is nothing to validate, and the merge has nothing to
		// apply.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual(bank.Name)
			response.JSON().Path("$.mask").String().IsEqual(*bank.Mask)
			response.JSON().Path("$.currency").String().IsEqual(bank.Currency)
			response.JSON().Path("$.availableBalance").Number().IsEqual(bank.AvailableBalance)
			response.JSON().Path("$.currentBalance").Number().IsEqual(bank.CurrentBalance)
			response.JSON().Path("$.status").String().IsEqual(string(bank.Status))
			response.JSON().Path("$.accountType").String().IsEqual(string(bank.AccountType))
			response.JSON().Path("$.accountSubType").String().IsEqual(string(bank.AccountSubType))
		}
	})

	t.Run("manual bank account rejects a null name", func(t *testing.T) {
		// Name is not a nullable field, so even though the key is optional an
		// explicit null should be rejected by the Required rule rather than
		// sneaking through validation and being silently ignored by the merge.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": nil,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.name").String().IsEqual("Name is required")
		}
	})

	t.Run("manual bank account rejects an empty name", func(t *testing.T) {
		// An empty string is not the same as the key being absent. The length rule
		// inside Name skips empty values, so without the Required rule an empty
		// name would slip through and blank out the account name. This proves it
		// does not.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.name").String().IsEqual("Name is required")
		}
	})

	t.Run("manual bank account rejects an empty currency", func(t *testing.T) {
		// Same empty string concern as the name. The currency format rules skip
		// empty values, so the Required rule is what stops an empty currency from
		// sneaking through.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currency": "",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("Currency is required")
		}
	})

	t.Run("manual bank account rejects a null balance", func(t *testing.T) {
		// Same idea as the null name, the balances are not nullable so an explicit
		// null should be rejected.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": nil,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.availableBalance").String().IsEqual("must not be nil")
		}
	})

	t.Run("manual bank account rejects an invalid mask", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // A mask that is not exactly four digits should trip the mask rule.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"mask": "12345",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			// Same structured one of envelope as the create path. We only care that
			// the four digit rule is the reason the real mask branch failed.
			response.JSON().Path("$.problems.mask.oneOf").Array().ContainsAll("Mask must be exactly 4 digits")
		}
	})

	t.Run("manual bank account rejects an invalid currency", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // A lower case currency passes the length and alpha rules but trips the
			// upper case rule.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currency": "usd",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be all upper case")
		}
	})

	t.Run("manual bank account rejects an unsupported currency", func(t *testing.T) {
		// A currency that is the right shape (three upper case letters) but is not
		// a currency the server actually knows about should trip the supported list
		// rule. This is the same case the put endpoint covers.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currency": "ZZZ",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be one supported by the server")
		}
	})

	t.Run("manual bank account rejects a name that is too long", func(t *testing.T) {
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // The name rule caps at 300 characters, so something longer should be
			// rejected.
			tooLong := ""
			for i := 0; i < 301; i++ {
				tooLong += "a"
			}
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": tooLong,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.name").String().IsEqual("Name must be between 1 and 300 characters")
		}
	})

	t.Run("plaid bank account can only patch the name", func(t *testing.T) {
		// A non-manual (Plaid) bank account uses the much more restrictive patch
		// schema that only allows the name to be changed. Plaid owns everything
		// else so the client should not be able to touch it.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveAPlaidBankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // The name can be changed.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": "My New Name",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual("My New Name")
			// Everything else should be untouched since it is not part of the schema.
			response.JSON().Path("$.mask").String().IsEqual(*bank.Mask)
			response.JSON().Path("$.availableBalance").Number().IsEqual(bank.AvailableBalance)
			response.JSON().Path("$.currentBalance").Number().IsEqual(bank.CurrentBalance)
			response.JSON().Path("$.status").String().IsEqual(string(bank.Status))
		}
	})

	t.Run("plaid bank account cannot patch mask currency or balances", func(t *testing.T) {
		// This is the whole reason there are two patch schemas. A Plaid bank
		// account is not allowed to change its mask, currency, or balances, so
		// those keys should come back as unexpected.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveAPlaidBankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // The mask is allowed on a manual account but NOT on a Plaid one.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"mask": "4321",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.mask").String().IsEqual("key not expected")
		}

		{ // Same for the currency.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currency": "EUR",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("key not expected")
		}

		{ // And the balances.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": -100,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.availableBalance").String().IsEqual("key not expected")
		}
	})

	t.Run("happy path patch a lunch flow bank account", func(t *testing.T) {
		// A lunch flow account is synced externally so monetr owns the balances,
		// but the user still picks the name, mask, and currency for it. It gets its
		// own patch schema (not the manual one and not the bare Plaid one) so it
		// can change those three things and nothing else. We route to it off of the
		// LunchFlowBankAccountId rather than the isManual check, which is what
		// these tests are really guarding.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveALunchFlowBankAccount(t, app.Clock, &link)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // The name, mask, and currency can all be changed.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name":     "My New Name",
					"mask":     "4321",
					"currency": "EUR",
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual("My New Name")
			response.JSON().Path("$.mask").String().IsEqual("4321")
			response.JSON().Path("$.currency").String().IsEqual("EUR")
		}

		{ // Read it back to make sure the patched values were actually persisted
			// and not just echoed back from the handler.
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.name").String().IsEqual("My New Name")
			response.JSON().Path("$.mask").String().IsEqual("4321")
			response.JSON().Path("$.currency").String().IsEqual("EUR")
		}
	})

	t.Run("lunch flow bank account cannot patch balances or classification", func(t *testing.T) {
		// The whole point of the dedicated lunch flow schema is that balances and
		// the account classification are synced externally, so the client must not
		// be able to touch them. These keys are not part of the schema so they
		// should come back as unexpected. This also guards against the routing ever
		// accidentally treating a lunch flow account as manual, which would let a
		// client change balances on an account that monetr keeps in sync.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveALunchFlowBankAccount(t, app.Clock, &link)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // The available balance is allowed on a manual account but NOT on a lunch
			// flow one.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": -100,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.availableBalance").String().IsEqual("key not expected")
		}

		{ // Same for the limit balance.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"limitBalance": 5000,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.limitBalance").String().IsEqual("key not expected")
		}

		{ // And the account type, monetr owns the classification for a synced
			// account.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"accountType": CreditBankAccountType,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.accountType").String().IsEqual("key not expected")
		}
	})

	t.Run("lunch flow bank account can clear the mask", func(t *testing.T) {
		// The mask field on the lunch flow patch schema is a one of null or a real
		// mask, just like the manual schema, so a client should be able to
		// explicitly null it out and actually have it cleared.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveALunchFlowBankAccount(t, app.Clock, &link)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // Make sure the mask starts off set.
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.mask").String().IsEqual(*bank.Mask)
		}

		{ // Then null it out.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"mask": nil,
				}).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.mask").IsNull()
		}

		{ // And make sure it actually persisted as null.
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.mask").IsNull()
		}
	})

	t.Run("lunch flow bank account rejects an invalid currency", func(t *testing.T) {
		// The currency goes through the same CurrencyCode rule as the manual
		// schema, so a lower case value trips the upper case rule and an
		// unsupported value trips the supported list rule.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveALunchFlowBankAccount(t, app.Clock, &link)

		token = GivenILogin(t, e, user.Login.Email, password)

		{ // A lower case currency passes the length and alpha rules but trips the
			// upper case rule.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currency": "usd",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be all upper case")
		}

		{ // A correctly shaped but unsupported currency trips the supported list
			// rule.
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"currency": "ZZZ",
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.currency").String().IsEqual("Currency must be one supported by the server")
		}
	})

	t.Run("lunch flow bank account rejects a null name", func(t *testing.T) {
		// Name is not nullable on the lunch flow schema either, so even though the
		// key is optional an explicit null should be rejected by the Required rule.
		app, e := NewTestApplication(t)
		var token string
		var bank BankAccount

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveALunchFlowLink(t, app.Clock, user)
		bank = fixtures.GivenIHaveALunchFlowBankAccount(t, app.Clock, &link)

		token = GivenILogin(t, e, user.Login.Email, password)

		{
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"name": nil,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").String().IsEqual("Invalid request")
			response.JSON().Path("$.problems.name").String().IsEqual("Name is required")
		}
	})

	t.Run("invalid bank account Id", func(t *testing.T) {
		// A value that is not even shaped like a bank account Id should be rejected
		// before we hit the repository at all.
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PATCH("/api/bank_accounts/{bankAccountId}").
			WithPath("bankAccountId", "potato").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]any{
				"name": "My New Name",
			}).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid bank account Id")
	})

	t.Run("cant patch someone elses bank account", func(t *testing.T) {
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

		{ // Try to patch the bank account
			response := e.PATCH("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bank.BankAccountId).
				WithCookie(TestCookieName, token).
				WithJSON(map[string]any{
					"availableBalance": -100,
				}).
				Expect()

			response.Status(http.StatusNotFound)
			response.JSON().Path("$.error").IsEqual("failed to retrieve bank account: record does not exist")
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

	t.Run("delete lunch flow bank account", func(t *testing.T) {
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
					"lunchFlowURL": "https://www.lunchflow.app/api/v1",
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

		var bankAccountId ID[BankAccount]
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
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
			bankAccountId = ID[BankAccount](response.JSON().Path("$.bankAccountId").String().Raw())
		}

		{ // Read the bank account back to make sure its in the right status
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.lunchFlowBankAccountId").String().IsEqual(string(lunchFlowBankAccountId))
			response.JSON().Path("$.lunchFlowBankAccount.lunchFlowBankAccountId").String().IsEqual(string(lunchFlowBankAccountId))
			response.JSON().Path("$.lunchFlowBankAccount.status").String().IsEqual(string(LunchFlowBankAccountStatusActive))
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusActive))
		}

		{ // Delete the actual bank account now
			response := e.DELETE("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.Body().IsEmpty()
		}

		{ // Read the bank account and make sure its deleted now!
			response := e.GET("/api/bank_accounts/{bankAccountId}").
				WithPath("bankAccountId", bankAccountId).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusOK)
			response.JSON().Path("$.bankAccountId").String().IsASCII().NotEmpty()
			response.JSON().Path("$.linkId").String().IsEqual(linkId.String())
			response.JSON().Path("$.lunchFlowBankAccountId").String().IsEqual(string(lunchFlowBankAccountId))
			response.JSON().Path("$.lunchFlowBankAccount.lunchFlowBankAccountId").String().IsEqual(string(lunchFlowBankAccountId))
			response.JSON().Path("$.lunchFlowBankAccount.status").String().IsEqual(string(LunchFlowBankAccountStatusInactive))
			response.JSON().Path("$.status").String().IsEqual(string(BankAccountStatusInactive))
		}
	})
}
