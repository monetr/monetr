package controller_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/internal/fixtures"
	. "github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// captureTransport is a sentry transport that just keeps every event it is
// given so that a test can make assertions about what was actually reported.
type captureTransport struct {
	lock   sync.Mutex
	events []*sentry.Event
}

func (t *captureTransport) Configure(_ sentry.ClientOptions)        {}
func (t *captureTransport) Flush(_ time.Duration) bool              { return true }
func (t *captureTransport) FlushWithContext(_ context.Context) bool { return true }
func (t *captureTransport) Close()                                  {}

func (t *captureTransport) SendEvent(event *sentry.Event) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.events = append(t.events, event)
}

func (t *captureTransport) Events() []*sentry.Event {
	t.lock.Lock()
	defer t.lock.Unlock()
	return append([]*sentry.Event(nil), t.events...)
}

// GivenSentryIsCapturing binds a client with a capturing transport to the
// current sentry hub. The middleware only records breadcrumbs onto the hub that
// sentryecho clones from the current one, so without a client bound here
// nothing would ever be reported and there would be nothing to assert against.
func GivenSentryIsCapturing(t *testing.T) *captureTransport {
	transport := &captureTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Transport:     transport,
		Integrations:  func([]sentry.Integration) []sentry.Integration { return nil },
		SampleRate:    1.0,
		DisableLogs:   true,
		EnableTracing: false,
	})
	require.NoError(t, err, "must be able to build the capturing sentry client")

	hub := sentry.CurrentHub()
	previous := hub.Client()
	hub.BindClient(client)
	t.Cleanup(func() {
		hub.BindClient(previous)
	})

	return transport
}

// authenticationBreadcrumbs pulls out the breadcrumbs that the authentication
// middleware is responsible for adding.
func authenticationBreadcrumbs(event *sentry.Event) []*sentry.Breadcrumb {
	crumbs := make([]*sentry.Breadcrumb, 0, len(event.Breadcrumbs))
	for _, crumb := range event.Breadcrumbs {
		if crumb.Category == "authentication" {
			crumbs = append(crumbs, crumb)
		}
	}
	return crumbs
}

// givenIHaveABankAccountAndToken seeds an account with a manual link and a bank
// account, and logs in, returning the bank account and the session token.
func givenIHaveABankAccountAndToken(t *testing.T, app *TestApp, e *httpexpect.Expect) (BankAccount, string) {
	user, password := fixtures.GivenIHaveABasicAccount(t, app.Clock)
	link := fixtures.GivenIHaveAManualLink(t, app.Clock, user)
	bank := fixtures.GivenIHaveABankAccount(t, app.Clock, &link, DepositoryBankAccountType, CheckingBankAccountSubType)
	token := GivenILogin(t, e, user.Login.Email, password)
	return bank, token
}

// TestAuthenticationBreadcrumbs covers the sentry breadcrumb that each of the
// authentication middlewares leaves behind. These breadcrumbs are only attached
// to an event if they were added to the hub's scope *before* the event was
// captured, so an upload that fails inside the handler is used to force an
// error to be reported after authentication has already happened.
func TestAuthenticationBreadcrumbs(t *testing.T) {
	// failTheUpload makes the file storage reject the upload, which makes the
	// handler report an error to sentry and return a 500.
	failTheUpload := func(app *TestApp) {
		app.Storage.EXPECT().
			Store(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(1).
			Return(errors.New("no space available"))
	}

	t.Run("api key auth is recorded before the handler runs", func(t *testing.T) {
		transport := GivenSentryIsCapturing(t)
		app, e := NewTestApplication(t)
		bank, token := givenIHaveABankAccountAndToken(t, app, e)
		apiKeyId, apiKeySecret := GivenIHaveAnApiKey(t, e, token)
		failTheUpload(app)

		e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithBasicAuth(apiKeyId, apiKeySecret).
			Expect().
			Status(http.StatusInternalServerError)

		events := transport.Events()
		require.NotEmpty(t, events, "the failed upload must have been reported to sentry")

		// The breadcrumb is added by a defer, so if that defer is registered in the
		// closure that calls next(ctx) then it does not run until after the handler
		// has already captured its event, and the event arrives with no
		// authentication breadcrumb at all.
		crumbs := authenticationBreadcrumbs(events[len(events)-1])
		require.Len(t, crumbs, 1, "the reported event must carry the authentication breadcrumb")
		assert.Equal(t, "Auth is valid", crumbs[0].Message, "a valid api key must not be reported as invalid auth")
		assert.Equal(t, "key", crumbs[0].Data["source"], "the auth source should be the api key")
	})

	t.Run("cookie auth is not reported as invalid", func(t *testing.T) {
		transport := GivenSentryIsCapturing(t)
		app, e := NewTestApplication(t)
		bank, token := givenIHaveABankAccountAndToken(t, app, e)
		failTheUpload(app)

		e.POST("/api/bank_accounts/{bankAccountId}/transactions/upload").
			WithPath("bankAccountId", bank.BankAccountId).
			WithMultipart().
			WithFileBytes("data", "transactions.ofx", fixtures.LoadFile(t, "sample-part-one.ofx")).
			WithCookie(TestCookieName, token).
			Expect().
			Status(http.StatusInternalServerError)

		events := transport.Events()
		require.NotEmpty(t, events, "the failed upload must have been reported to sentry")

		crumbs := authenticationBreadcrumbs(events[len(events)-1])
		require.Len(t, crumbs, 1, "the reported event must carry the authentication breadcrumb")
		assert.Equal(t, "Auth is valid", crumbs[0].Message, "a valid session cookie must not be reported as invalid auth")
		assert.Equal(t, "cookie", crumbs[0].Data["source"], "the auth source should be the cookie")
	})
}
