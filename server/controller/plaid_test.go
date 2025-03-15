package controller_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mock_plaid"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/plaid/plaid-go/v30/plaid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostTokenCallback(t *testing.T) {
	t.Run("cant retrieve accounts", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		publicToken := mock_plaid.MockExchangePublicToken(t)
		mock_plaid.MockGetAccounts(t, nil)

		response := e.POST("/api/plaid/link/token/callback").
			WithCookie(TestCookieName, token).
			WithJSON(map[string]interface{}{
				"publicToken":     publicToken,
				"institutionId":   "123",
				"institutionName": gofakeit.Company(),
				"accountIds": []string{
					gofakeit.UUID(),
				},
			}).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().IsEqual("could not retrieve details for any accounts")
	})
}

func TestPutUpdatePlaidLink(t *testing.T) {
	t.Run("successful with account select enabled", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		mock_plaid.MockCreateLinkToken(t, func(t *testing.T, request plaid.LinkTokenCreateRequest) {
			assert.NotNil(t, request.GetUpdate().AccountSelectionEnabled, "account selection enabled cannot be nil")
			assert.True(t, *request.GetUpdate().AccountSelectionEnabled, "account selection enabled must be true")
		})

		response := e.PUT("/api/plaid/link/update/{linkId}").
			WithPath("linkId", link.LinkId).
			WithQuery("update_account_selection", "true").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.linkToken").String().NotEmpty()
		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"POST https://sandbox.plaid.com/link/token/create": 1,
		}, "must match expected Plaid API calls")
	})

	t.Run("successful with account select disabled", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		mock_plaid.MockCreateLinkToken(t, func(t *testing.T, request plaid.LinkTokenCreateRequest) {
			assert.NotNil(t, request.GetUpdate().AccountSelectionEnabled, "account selection enabled cannot be nil")
			assert.False(t, *request.GetUpdate().AccountSelectionEnabled, "account selection enabled must be false")
		})

		response := e.PUT("/api/plaid/link/update/{linkId}").
			WithPath("linkId", link.LinkId).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusOK)
		response.JSON().Path("$.linkToken").String().NotEmpty()
		assert.EqualValues(t, httpmock.GetCallCountInfo(), map[string]int{
			"POST https://sandbox.plaid.com/link/token/create": 1,
		}, "must match expected Plaid API calls")
	})

	t.Run("manual link", func(t *testing.T) {
		app, e := NewTestApplication(t)
		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.PUT("/api/plaid/link/update/{linkId}").
			WithPath("linkId", link.LinkId).
			WithQuery("update_account_selection", "true").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("cannot update a non-Plaid link")
	})

	t.Run("missing link ID", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT("/api/plaid/link/update/-1").
			WithQuery("update_account_selection", "true").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid link Id")
	})

	t.Run("bad link ID", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT("/api/plaid/link/update/0").
			WithQuery("update_account_selection", "true").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("must specify a valid link Id")
	})

	t.Run("missing link", func(t *testing.T) {
		_, e := NewTestApplication(t)
		token := GivenIHaveToken(t, e)

		response := e.PUT("/api/plaid/link/update/link_bogus").
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to retrieve link: record does not exist")
	})
}

func TestPostSyncPlaidManually(t *testing.T) {
	t.Run("successful with account select enabled", func(t *testing.T) {
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.SyncPlaid),
				testutils.NewGenericMatcher(func(args background.SyncPlaidArguments) bool {
					a := assert.EqualValues(t, link.LinkId, args.LinkId, "Link ID should match")
					b := assert.EqualValues(t, link.AccountId, args.AccountId, "Account ID should match")
					return a && b
				}),
			).
			MaxTimes(1).
			Return(nil)

		response := e.POST("/api/plaid/link/sync").
			WithJSON(map[string]interface{}{
				"linkId": link.LinkId,
			}).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusAccepted)
	})

	t.Run("fails on subsequent attempt", func(t *testing.T) {
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.SyncPlaid),
				testutils.NewGenericMatcher(func(args background.SyncPlaidArguments) bool {
					a := assert.EqualValues(t, link.LinkId, args.LinkId, "Link ID should match")
					b := assert.EqualValues(t, link.AccountId, args.AccountId, "Account ID should match")
					return a && b
				}),
			).
			MaxTimes(1).
			Return(nil)

		{ // First request should succeed.
			response := e.POST("/api/plaid/link/sync").
				WithJSON(map[string]interface{}{
					"linkId": link.LinkId,
				}).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusAccepted)
		}

		{ // Second request should fail, its too soon.
			response := e.POST("/api/plaid/link/sync").
				WithJSON(map[string]interface{}{
					"linkId": link.LinkId,
				}).
				WithCookie(TestCookieName, token).
				Expect()

			response.Status(http.StatusTooEarly)
			response.JSON().Path("$.error").String().IsEqual("link has been manually synced too recently")
		}
	})

	t.Run("failed to enque job", func(t *testing.T) {
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAPlaidLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		app.Jobs.EXPECT().
			EnqueueJob(
				gomock.Any(),
				gomock.Eq(background.SyncPlaid),
				testutils.NewGenericMatcher(func(args background.SyncPlaidArguments) bool {
					a := assert.EqualValues(t, link.LinkId, args.LinkId, "Link ID should match")
					b := assert.EqualValues(t, link.AccountId, args.AccountId, "Account ID should match")
					return a && b
				}),
			).
			MaxTimes(1).
			Return(errors.New("queue is offline"))

		response := e.POST("/api/plaid/link/sync").
			WithJSON(map[string]interface{}{
				"linkId": link.LinkId,
			}).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusInternalServerError)
		response.JSON().Path("$.error").String().IsEqual("failed to trigger manual sync")
	})

	t.Run("invalid link ID", func(t *testing.T) {
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/plaid/link/sync").
			WithJSON(map[string]interface{}{
				"linkId": "link_bogus",
			}).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusNotFound)
		response.JSON().Path("$.error").String().IsEqual("failed to retrieve link: record does not exist")
	})

	t.Run("manual link", func(t *testing.T) {
		app, e := NewTestApplication(t)

		user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
		link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
		token := GivenILogin(t, e, user.Login.Email, password)

		response := e.POST("/api/plaid/link/sync").
			WithJSON(map[string]interface{}{
				"linkId": link.LinkId,
			}).
			WithCookie(TestCookieName, token).
			Expect()

		response.Status(http.StatusBadRequest)
		response.JSON().Path("$.error").String().IsEqual("cannot manually sync a non-Plaid link")
	})
}
