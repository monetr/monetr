package controller_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/golang/mock/gomock"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/application"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/controller"
	"github.com/monetr/monetr/server/internal/mock_secrets"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/plaid/plaid-go/v14/plaid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	FifthteenthAndLastDayOfEveryMonth = "DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1"
	FirstDayOfEveryMonth              = "DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1"
)

const (
	TestEmailDomain   = "monetr.mini"
	TestUIDomainName  = "app.monetr.mini"
	TestAPIDomainName = "api.monetr.mini"
	TestCookieName    = "M-Token"
)

func NewTestApplicationConfig(t *testing.T) config.Configuration {
	return config.Configuration{
		UIDomainName:        TestUIDomainName,
		APIDomainName:       TestAPIDomainName,
		AllowSignUp:         true,
		ExternalURLProtocol: "https",
		Server: config.Server{
			Cookies: config.Cookies{
				SameSiteStrict: true,
				Secure:         true,
				Name:           TestCookieName,
			},
		},
		PostgreSQL: config.PostgreSQL{},
		Email: config.Email{
			Enabled: false,
			Verification: config.EmailVerification{
				Enabled:       false,
				TokenLifetime: 10 * time.Second,
				TokenSecret:   gofakeit.Generate("????????????????"),
			},
			Domain: TestEmailDomain,
			SMTP:   config.SMTPClient{},
		},
		ReCAPTCHA: config.ReCAPTCHA{},
		Plaid: config.Plaid{
			Enabled:      true,
			ClientID:     gofakeit.UUID(),
			ClientSecret: gofakeit.UUID(),
			Environment:  plaid.Sandbox,
		},
		CORS: config.CORS{
			Debug: false,
		},
		Logging: config.Logging{
			Level: "trace",
		},
	}
}

func NewTestApplication(t *testing.T) (*TestApp, *httpexpect.Expect) {
	configuration := NewTestApplicationConfig(t)
	return NewTestApplicationWithConfig(t, configuration)
}

type TestApp struct {
	Configuration config.Configuration
	Email         *mockgen.MockEmailCommunication
	Clock         *clock.Mock
	Tokens        security.ClientTokens
}

type TestAppInterfaces struct {
	JobController *background.JobController
}

func NewTestApplicationWithConfig(t *testing.T, configuration config.Configuration) (*TestApp, *httpexpect.Expect) {
	return NewTestApplicationPatched(t, configuration, TestAppInterfaces{})
}

func NewTestApplicationPatched(t *testing.T, configuration config.Configuration, patched TestAppInterfaces) (*TestApp, *httpexpect.Expect) {
	clock := clock.NewMock()
	clock.Set(time.Date(2023, 10, 9, 13, 32, 0, 0, time.UTC))
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)
	secretProvider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)
	plaidRepo := repository.NewPlaidRepository(db)
	plaidClient := platypus.NewPlaid(log, secretProvider, plaidRepo, configuration.Plaid)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "must be able to generate keys")

	clientTokens, err := security.NewPasetoClientTokens(log, clock, configuration.APIDomainName, publicKey, privateKey)
	require.NoError(t, err, "must be able to init the client tokens interface")

	miniRedis := miniredis.NewMiniRedis()
	require.NoError(t, miniRedis.Start())
	redisPool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", miniRedis.Server().Addr().String())
		},
	}

	t.Cleanup(func() {
		require.NoError(t, redisPool.Close())
		miniRedis.Close()
	})
	plaidSecrets := mock_secrets.NewMockPlaidSecrets()

	var jobRunner background.JobController
	if patched.JobController != nil {
		jobRunner = *patched.JobController
	} else {
		jobRunner = background.NewSynchronousJobRunner(t, clock, plaidClient, plaidSecrets)
	}

	emailMockController := gomock.NewController(t)
	t.Cleanup(func() {
		defer emailMockController.Finish()
	})
	email := mockgen.NewMockEmailCommunication(emailMockController)

	c := controller.NewController(
		log,
		configuration,
		db,
		jobRunner,
		plaidClient,
		nil,
		stripe_helper.NewStripeHelper(log, gofakeit.UUID()),
		redisPool,
		plaidSecrets,
		billing.NewBasicPaywall(
			log,
			clock,
			billing.NewAccountRepository(log, cache.NewCache(log, redisPool), db),
		),
		email,
		clientTokens,
		clock,
	)
	app := application.NewApp(configuration, c)

	// run server using httptest
	server := httptest.NewServer(app)
	t.Cleanup(func() {
		require.NoError(t, c.Close(), "must be able to shutdown the monetr http controller")
		server.Close()
	})

	expect := httpexpect.WithConfig(httpexpect.Config{
		TestName: t.Name(),
		Client:   server.Client(),
		BaseURL:  server.URL,
		AssertionHandler: &httpexpect.DefaultAssertionHandler{
			Formatter: &httpexpect.DefaultFormatter{
				DisableNames:     false,
				DisablePaths:     false,
				DisableAliases:   false,
				DisableDiffs:     false,
				DisableRequests:  false,
				DisableResponses: false,
				DigitSeparator:   httpexpect.DigitSeparatorComma,
				FloatFormat:      httpexpect.FloatFormatAuto,
				StacktraceMode:   httpexpect.StacktraceModeStandard,
				ColorMode:        httpexpect.ColorModeAuto,
			},
			Reporter: httpexpect.NewAssertReporter(t),
		},

		Printers: []httpexpect.Printer{
			NewDebugPrinter(log, true),
		},
		// Reporter: httpexpect.NewAssertReporter(t),
		// Formatter: ,
		Context: context.WithValue(context.Background(), "test", t.Name()),
	})

	return &TestApp{
		Configuration: configuration,
		Email:         email,
		Clock:         clock,
		Tokens:        clientTokens,
	}, expect
}

type DebugPrinter struct {
	logger *logrus.Entry
	body   bool
}

// NewDebugPrinter returns a new DebugPrinter given a logger and body
// flag. If body is true, request and response body is also printed.
func NewDebugPrinter(logger *logrus.Entry, body bool) DebugPrinter {
	return DebugPrinter{logger, body}
}

// Request implements Printer.Request.
func (p DebugPrinter) Request(req *http.Request) {
	if req == nil {
		return
	}

	dump, err := httputil.DumpRequest(req, p.body)
	if err != nil {
		panic(err)
	}
	p.logger.Debug("Logging Request\n" + string(dump) + "\n\t")
}

// Response implements Printer.Response.
func (p DebugPrinter) Response(resp *http.Response, duration time.Duration) {
	if resp == nil {
		return
	}

	dump, err := httputil.DumpResponse(resp, p.body)
	if err != nil {
		panic(err)
	}

	text := strings.Replace(string(dump), "\r\n", "\n", -1)
	lines := strings.SplitN(text, "\n", 2)

	p.logger.Debugf("Logging Response\n%s %s\n%s\t", lines[0], duration, lines[1])
}

func GivenIHaveToken(t *testing.T, e *httpexpect.Expect) string {
	email, password := register(t, e)
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	loginRequest.Email = email
	loginRequest.Password = password
	var token string
	{
		response := e.POST("/api/authentication/login").
			WithJSON(loginRequest).
			Expect()

		response.Status(http.StatusOK)
		response.Cookie(TestCookieName).Value().NotEmpty()
		token = response.Cookie(TestCookieName).Value().Raw()
	}
	require.NotEmpty(t, token, "token from login must not be empty")
	return token
}

func register(t *testing.T, e *httpexpect.Expect) (email, password string) {
	var registerRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	registerRequest.Email = testutils.GetUniqueEmail(t)
	registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
	registerRequest.FirstName = gofakeit.FirstName()
	registerRequest.LastName = gofakeit.LastName()

	response := e.POST(`/api/authentication/register`).
		WithJSON(registerRequest).
		Expect()

	response.Status(http.StatusOK)
	return registerRequest.Email, registerRequest.Password
}

func GivenIHaveLogin(t *testing.T, e *httpexpect.Expect) (email, password string) {
	email, password = register(t, e)
	require.NotEmpty(t, email, "email cannot be empty")
	require.NotEmpty(t, password, "password cannot be empty")
	return
}

func GivenILogin(t *testing.T, e *httpexpect.Expect, email, password string) (token string) {
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	loginRequest.Email = email
	loginRequest.Password = password

	response := e.POST(`/api/authentication/login`).
		WithJSON(loginRequest).
		Expect()

	response.Status(http.StatusOK)
	AssertSetTokenCookie(t, response)
	return response.Cookie(TestCookieName).Value().Raw()
}

func AssertSetTokenCookie(t *testing.T, response *httpexpect.Response) string {
	cookie := response.Cookie(TestCookieName)
	require.NotNil(t, cookie, "auth cookie must not be nil if they were authenticated")
	cookie.Path().IsEqual("/")
	cookie.Domain().IsEqual(TestAPIDomainName)
	raw := cookie.Raw()
	require.NotNil(t, raw, "raw cookie must not be nil if authentication was successful, or you werent authenticated")
	assert.True(t, raw.Secure, "cookie must be secure")
	assert.True(t, raw.HttpOnly, "cookie must be secure")

	// This assertion is here to prevent a regression. We want to make sure that requests that would previously
	// return a token in the body, do not anymore.
	response.JSON().Object().NotContainsKey("token")
	return cookie.Value().Raw()
}

func MustSendVerificationEmail(t *testing.T, app *TestApp, n int) {
	app.Email.
		EXPECT().
		SendVerification(
			gomock.Any(),
			gomock.Any(),
		).
		Return(nil).
		Times(n).
		Do(func(ctx context.Context, params communication.VerifyEmailParams) error {
			require.NotNil(t, ctx, "email context cannot be nil")
			require.NotEmpty(t, params.Email, "verification email address cannot be empty")
			require.NotEmpty(t, params.FirstName, "verification email first name cannot be empty")
			require.NotEmpty(t, params.LastName, "verification email last name cannot be empty")
			require.NotEmpty(t, params.BaseURL, "verification email base url must be defined")
			require.NotEmpty(t, params.VerifyURL, "verification email verify url must be defined")
			return nil
		})
}

func MustSendPasswordResetEmail(t *testing.T, app *TestApp, n int, emails ...string) {
	app.Email.
		EXPECT().
		SendPasswordReset(
			gomock.Any(),
			gomock.Any(),
		).
		Return(nil).
		Times(n).
		Do(func(ctx context.Context, params communication.PasswordResetParams) error {
			require.NotNil(t, ctx, "email context cannot be nil")
			require.NotEmpty(t, params.Email, "password reset email address cannot be empty")
			require.NotEmpty(t, params.FirstName, "password reset email first name cannot be empty")
			require.NotEmpty(t, params.LastName, "password reset email last name cannot be empty")
			require.NotEmpty(t, params.BaseURL, "password reset email base url must be defined")
			require.NotEmpty(t, params.ResetURL, "password reset email url must be defined")
			if len(emails) > 0 {
				for _, email := range emails {
					if strings.EqualFold(email, params.Email) {
						return nil
					}
				}
				// If none of the emails match then something is wrong.
				t.Fatalf(
					"email specified for reset password <%s> was not expected, expected address(es): %s",
					params.Email,
					strings.Join(emails, ", "),
				)
			}
			return nil
		})
}
