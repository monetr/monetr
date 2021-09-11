package controller_test

import (
	"net/http"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gomodule/redigo/redis"
	"github.com/kataras/iris/v12/httptest"
	"github.com/monetr/rest-api/pkg/application"
	"github.com/monetr/rest-api/pkg/billing"
	"github.com/monetr/rest-api/pkg/cache"
	"github.com/monetr/rest-api/pkg/config"
	"github.com/monetr/rest-api/pkg/controller"
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
		UIDomainName:  "http://localhost:1234",
		APIDomainName: "http://localhost:1235",
		AllowSignUp:   true,
		JWT: config.JWT{
			LoginJwtSecret:        gofakeit.UUID(),
			RegistrationJwtSecret: gofakeit.UUID(),
		},
		PostgreSQL: config.PostgreSQL{},
		EMail:      config.Email{},
		ReCAPTCHA:  config.ReCAPTCHA{},
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

func NewTestApplicationWithConfig(t *testing.T, configuration config.Configuration) *httptest.Expect {
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
	)
	app := application.NewApp(configuration, c)
	return httptest.New(t, app)
}

func GivenIHaveToken(t *testing.T, e *httptest.Expect) string {
	_, _, token := register(t, e)
	return token
}

func register(t *testing.T, e *httptest.Expect) (email, password, token string) {
	var registerRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}
	registerRequest.Email = testutils.GivenIHaveAnEmail(t)
	registerRequest.Password = gofakeit.Password(true, true, true, true, false, 32)
	registerRequest.FirstName = gofakeit.FirstName()
	registerRequest.LastName = gofakeit.LastName()

	response := e.POST(`/authentication/register`).
		WithJSON(registerRequest).
		Expect()

	response.Status(http.StatusOK)
	token = response.JSON().Path("$.token").String().Raw()
	require.NotEmpty(t, token, "token cannot be empty")

	return registerRequest.Email, registerRequest.Password, token
}

func GivenIHaveLogin(t *testing.T, e *httptest.Expect) (email, password string) {
	email, password, _ = register(t, e)
	return
}
