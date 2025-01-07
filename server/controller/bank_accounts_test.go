package controller_test

import (
	"net/http"
	"testing"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/myownsanity"
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
				WithJSON(Link{
					LinkType:        ManualLinkType,
					InstitutionName: "Manual Link",
					Description:     myownsanity.StringP("My personal link"),
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
				WithJSON(BankAccount{
					LinkId:           linkId,
					AvailableBalance: 100,
					CurrentBalance:   100,
					LimitBalance:     0,
					Mask:             "1234",
					Name:             "Checking Account",
					OriginalName:     "PERSONAL CHECKING",
					Type:             DepositoryBankAccountType,
					SubType:          CheckingBankAccountSubType,
					Status:           ActiveBankAccountStatus,
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

	t.Run("requires a link", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		{ // Create the manual bank account
			response := e.POST("/api/bank_accounts").
				WithCookie(TestCookieName, token).
				WithJSON(BankAccount{
					LinkId:           "bogus",
					AvailableBalance: 100,
					CurrentBalance:   100,
					LimitBalance:     0,
					Mask:             "1234",
					Name:             "Checking Account",
					OriginalName:     "PERSONAL CHECKING",
					Type:             DepositoryBankAccountType,
					SubType:          CheckingBankAccountSubType,
					Status:           ActiveBankAccountStatus,
				}).
				Expect()

			// This returns the same error as if you provide a valid link, because it
			// just sees that the link does not exist with a manual type. Not that the
			// link doesn't exist at all.
			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Cannot create a bank account for a non-manual link")
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
				WithJSON(BankAccount{
					LinkId:           plaidLink.LinkId,
					AvailableBalance: 100,
					CurrentBalance:   100,
					LimitBalance:     0,
					Mask:             "1234",
					Name:             "Checking Account",
					OriginalName:     "PERSONAL CHECKING",
					Type:             DepositoryBankAccountType,
					SubType:          CheckingBankAccountSubType,
					Status:           ActiveBankAccountStatus,
				}).
				Expect()

			response.Status(http.StatusBadRequest)
			response.JSON().Path("$.error").IsEqual("Cannot create a bank account for a non-manual link")
		}
	})
}
