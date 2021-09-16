package controller_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gomodule/redigo/redis"
	"github.com/kataras/iris/v12/httptest"
	"github.com/monetr/rest-api/pkg/application"
	"github.com/monetr/rest-api/pkg/billing"
	"github.com/monetr/rest-api/pkg/cache"
	"github.com/monetr/rest-api/pkg/config"
	"github.com/monetr/rest-api/pkg/controller"
	"github.com/monetr/rest-api/pkg/internal/mock_mail"
	"github.com/monetr/rest-api/pkg/internal/mock_secrets"
	"github.com/monetr/rest-api/pkg/internal/platypus"
	"github.com/monetr/rest-api/pkg/internal/stripe_helper"
	"github.com/monetr/rest-api/pkg/internal/testutils"
	"github.com/monetr/rest-api/pkg/jobs"
	"github.com/monetr/rest-api/pkg/repository"
	"github.com/monetr/rest-api/pkg/secrets"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
)

func NewTestApplicationConfig(t *testing.T) config.Configuration {
	return config.Configuration{
		Name:          t.Name(),
		UIDomainName:  "ui.monetr.mini",
		APIDomainName: "api.monetr.mini",
		AllowSignUp:   true,
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
			Domain: "monetr.mini",
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

func NewTestApplication(t *testing.T) *httptest.Expect {
	configuration := NewTestApplicationConfig(t)
	return NewTestApplicationWithConfig(t, configuration)
}

type TestApp struct {
	Mail *mock_mail.MockMailCommunication
}

func NewTestApplicationExWithConfig(t *testing.T, configuration config.Configuration) (*TestApp, *httptest.Expect) {
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

	mockJobManager := jobs.NewNonDistributedJobManager(
		log,
		redisPool,
		db,
		plaidClient,
		nil,
		plaidSecrets,
	)

	mockMail := mock_mail.NewMockMail()

	c := controller.NewController(
		log,
		configuration,
		db,
		mockJobManager,
		plaidClient,
		nil,
		stripe_helper.NewStripeHelper(log, gofakeit.UUID()),
		redisPool,
		plaidSecrets,
		billing.NewBasicPaywall(log, billing.NewAccountRepository(log, cache.NewCache(log, redisPool), db)),
		mockMail,
	)
	app := application.NewApp(configuration, c)
	return &TestApp{
		Mail: mockMail,
	}, httptest.New(t, app)
}

func NewTestApplicationWithConfig(t *testing.T, configuration config.Configuration) *httptest.Expect {
	_, e := NewTestApplicationExWithConfig(t, configuration)
	return e
}

func GivenIHaveToken(t *testing.T, e *httptest.Expect) string {
	_, _, result := register(t, e)
	require.Contains(t, result, "token", "result must contain token")
	require.IsType(t, string(""), result["token"], "token must be a string")
	return result["token"].(string)
}

func register(t *testing.T, e *httptest.Expect) (email, password string, result map[string]interface{}) {
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

	response := e.POST(`/authentication/register`).
		WithJSON(registerRequest).
		Expect()

	response.Status(http.StatusOK)
	return registerRequest.Email, registerRequest.Password, response.JSON().Object().Raw()
}

func GivenIHaveLogin(t *testing.T, e *httptest.Expect) (email, password string) {
	email, password, _ = register(t, e)
	require.NotEmpty(t, email, "email cannot be empty")
	require.NotEmpty(t, password, "password cannot be empty")
	return
}
