package controller_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/pkg/application"
	"github.com/monetr/monetr/pkg/background"
	"github.com/monetr/monetr/pkg/billing"
	"github.com/monetr/monetr/pkg/cache"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/controller"
	"github.com/monetr/monetr/pkg/internal/mock_mail"
	"github.com/monetr/monetr/pkg/internal/mock_secrets"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/platypus"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/monetr/monetr/pkg/secrets"
	"github.com/monetr/monetr/pkg/stripe_helper"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestEmailDomain   = "monetr.mini"
	TestUIDomainName  = "app.monetr.mini"
	TestAPIDomainName = "api.monetr.mini"
	TestCookieName    = "M-Token"
)

func NewTestApplicationConfig(t *testing.T) config.Configuration {
	return config.Configuration{
		Name:          t.Name(),
		UIDomainName:  TestUIDomainName,
		APIDomainName: TestAPIDomainName,
		AllowSignUp:   true,
		Server: config.Server{
			Cookies: config.Cookies{
				SameSiteStrict: true,
				Secure:         true,
				Name:           TestCookieName,
			},
		},
		JWT: config.JWT{
			LoginJwtSecret:        gofakeit.UUID(),
			RegistrationJwtSecret: gofakeit.UUID(),
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

func NewTestApplication(t *testing.T) *httpexpect.Expect {
	configuration := NewTestApplicationConfig(t)
	return NewTestApplicationWithConfig(t, configuration)
}

type TestApp struct {
	Mail *mock_mail.MockMailCommunication
}

func NewTestApplicationExWithConfig(t *testing.T, configuration config.Configuration) (*TestApp, *httpexpect.Expect) {
	log := testutils.GetLog(t)
	db := testutils.GetPgDatabase(t)
	secretProvider := secrets.NewPostgresPlaidSecretsProvider(log, db)
	plaidRepo := repository.NewPlaidRepository(db)
	plaidClient := platypus.NewPlaid(log, secretProvider, plaidRepo, configuration.Plaid)

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

	jobRunner := background.NewSynchronousJobRunner(t, plaidClient, plaidSecrets)

	mockMail := mock_mail.NewMockMail()

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
		billing.NewBasicPaywall(log, billing.NewAccountRepository(log, cache.NewCache(log, redisPool), db)),
		mockMail,
	)
	app := application.NewApp(configuration, c)

	require.NoError(t, app.Build(), "must build app")

	// run server using httptest
	server := httptest.NewServer(app)
	t.Cleanup(func() {
		server.Close()
	})

	expect := httpexpect.WithConfig(httpexpect.Config{
		Client:   server.Client(),
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
		Printers: []httpexpect.Printer{},
		Context:  context.WithValue(context.Background(), "test", t.Name()),
	})

	return &TestApp{
		Mail: mockMail,
	}, expect
}

func NewTestApplicationWithConfig(t *testing.T, configuration config.Configuration) *httpexpect.Expect {
	_, e := NewTestApplicationExWithConfig(t, configuration)
	return e
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
	response.Cookie(TestCookieName).Path().Equal("/")
	response.Cookie(TestCookieName).Domain().Equal(TestAPIDomainName)
	assert.True(t, response.Cookie(TestCookieName).Raw().Secure, "cookie must be secure")
	assert.True(t, response.Cookie(TestCookieName).Raw().HttpOnly, "cookie must be secure")

	// This assertion is here to prevent a regression. We want to make sure that requests that would previously
	// return a token in the body, do not anymore.
	response.JSON().Object().NotContainsKey("token")
	return response.Cookie(TestCookieName).Value().Raw()
}
