package controller_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/application"
	"github.com/monetr/monetr/server/background"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/captcha"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/controller"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/platypus"
	"github.com/monetr/monetr/server/pubsub"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/monetr/monetr/server/security"
	"github.com/monetr/monetr/server/storage"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestJsonDecode(t *testing.T) {
	t.Run("handle merging json data", func(t *testing.T) {
		// This might be a really useless test, but I want something that concretely
		// proves that json decoding into a struct with existing values does not
		// trample all of the values of that struct. Instead it just updates the
		// values provided in the actual json.
		type Foo struct {
			Id        string `json:"id"`
			Name      string `json:"name"`
			Amount    int64  `json:"amount"`
			Nullable  *int   `json:"nullable"`
			Overwrite *int   `json:"overwrite"`
		}
		input := `{"name":"foobar", "overwrite": 5678}`
		nullable := 12345
		existing := Foo{
			Id:        "foo_1234",
			Name:      "oldname",
			Amount:    100,
			Nullable:  &nullable,
			Overwrite: &nullable,
		}
		err := json.Unmarshal([]byte(input), &existing)
		assert.NoError(t, err)
		assert.EqualValues(t, "foo_1234", existing.Id)
		assert.EqualValues(t, "foobar", existing.Name) // Changed
		assert.EqualValues(t, 100, existing.Amount)
		assert.EqualValues(t, nullable, *existing.Nullable)
		assert.EqualValues(t, 5678, *existing.Overwrite)
	})
}

const (
	FifthteenthAndLastDayOfEveryMonth = "DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1"
	FirstDayOfEveryMonth              = "DTSTART:20220101T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1"
)

const (
	TestCookieName = "M-Token"
)

func NewTestApplicationConfig(t *testing.T) config.Configuration {
	return testutils.GetConfig(t)
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
	kms := secrets.NewPlaintextKMS()
	plaidClient := platypus.NewPlaid(log, clock, kms, db, configuration.Plaid)

	plaidWebhooks := platypus.NewInMemoryWebhookVerification(
		log,
		plaidClient,
		1*time.Hour,
	)

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "must be able to generate keys")

	clientTokens, err := security.NewPasetoClientTokens(
		log,
		clock,
		configuration.Server.GetBaseURL().String(),
		publicKey,
		privateKey,
	)
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

	// Create a temporary directory for file uploads.
	tempDirectory, err := os.MkdirTemp("", fmt.Sprintf("monetr-test-%x", t.Name()))
	require.NoError(t, err, "must be able to create a temp directory for uploads")
	log.Debugf("[TEST] created temporary directory for uploads: %s", tempDirectory)

	fileStorage, err := storage.NewFilesystemStorage(log, tempDirectory)
	require.NoError(t, err, "must not have an error when creating the filesystem storage")

	var jobRunner background.JobController
	if patched.JobController != nil {
		jobRunner = *patched.JobController
	}

	emailMockController := gomock.NewController(t)
	t.Cleanup(func() {
		defer emailMockController.Finish()
	})
	email := mockgen.NewMockEmailCommunication(emailMockController)

	var recaptcha captcha.Verification
	if configuration.ReCAPTCHA.Enabled {
		recaptcha, err = captcha.NewReCAPTCHAVerification(
			configuration.ReCAPTCHA.PrivateKey,
		)
		if err != nil {
			panic(err)
		}
	}

	cachePool := cache.NewCache(log, redisPool)
	accountsRepo := repository.NewAccountRepository(log, cachePool, db)
	stripeHelper := stripe_helper.NewStripeHelper(log, gofakeit.UUID())
	pubSub := pubsub.NewPostgresPubSub(log, db)
	plaidInstitutions := platypus.NewPlaidInstitutionWrapper(
		log,
		plaidClient,
		cachePool,
	)

	bill := billing.NewBilling(
		log,
		clock,
		configuration,
		accountsRepo,
		stripeHelper,
		pubSub,
	)

	c := &controller.Controller{
		Accounts:                 accountsRepo,
		Billing:                  bill,
		Cache:                    cachePool,
		Captcha:                  recaptcha,
		ClientTokens:             clientTokens,
		Clock:                    clock,
		Configuration:            configuration,
		DB:                       db,
		Email:                    email,
		FileStorage:              fileStorage,
		JobRunner:                jobRunner,
		KMS:                      kms,
		Log:                      log,
		Plaid:                    plaidClient,
		PlaidInstitutions:        plaidInstitutions,
		PlaidWebhookVerification: plaidWebhooks,
		PubSub:                   pubSub,
		Stats:                    nil,
		Stripe:                   stripeHelper,
	}

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
			testutils.NewDebugPrinter(log, true),
		},
		Context: context.WithValue(context.Background(), "test", t.Name()),
	})

	return &TestApp{
		Configuration: configuration,
		Email:         email,
		Clock:         clock,
		Tokens:        clientTokens,
	}, expect
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
		Locale    string `json:"locale"`
		Timezone  string `json:"timezone"`
	}
	registerRequest.Email = testutils.GetUniqueEmail(t)
	registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
	registerRequest.FirstName = gofakeit.FirstName()
	registerRequest.LastName = gofakeit.LastName()
	registerRequest.Locale = "en_US"
	registerRequest.Timezone = "America/Chicago"

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
	cookie.Domain().IsEqual("monetr.local")
	raw := cookie.Raw()
	require.NotNil(t, raw, "raw cookie must not be nil if authentication was successful, or you werent authenticated")
	assert.True(t, raw.Secure, "cookie must be secure")
	assert.True(t, raw.HttpOnly, "cookie should always be http only")

	// This assertion is here to prevent a regression. We want to make sure that
	// requests that would previously return a token in the body, do not anymore.
	response.JSON().Object().NotContainsKey("token")
	return cookie.Value().Raw()
}

func MustSendVerificationEmail(t *testing.T, app *TestApp, n int) {
	app.Email.
		EXPECT().
		SendEmail(
			gomock.Any(),
			gomock.AssignableToTypeOf(communication.VerifyEmailParams{}),
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
		SendEmail(
			gomock.Any(),
			gomock.AssignableToTypeOf(communication.PasswordResetParams{}),
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

func MustSendPasswordChangedEmail(t *testing.T, app *TestApp, n int, emails ...string) {
	app.Email.
		EXPECT().
		SendEmail(
			gomock.Any(),
			gomock.AssignableToTypeOf(communication.PasswordChangedParams{}),
		).
		Return(nil).
		Times(n).
		Do(func(ctx context.Context, params communication.PasswordChangedParams) error {
			require.NotNil(t, ctx, "email context cannot be nil")
			require.NotEmpty(t, params.Email, "password changed email address cannot be empty")
			require.NotEmpty(t, params.FirstName, "password changed email first name cannot be empty")
			require.NotEmpty(t, params.LastName, "password changed email last name cannot be empty")
			require.NotEmpty(t, params.BaseURL, "password changed email base url must be defined")
			if len(emails) > 0 {
				for _, email := range emails {
					if strings.EqualFold(email, params.Email) {
						return nil
					}
				}
				// If none of the emails match then something is wrong.
				t.Fatalf(
					"email specified for password changed <%s> was not expected, expected address(es): %s",
					params.Email,
					strings.Join(emails, ", "),
				)
			}
			return nil
		})
}
