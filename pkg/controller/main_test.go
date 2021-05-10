package controller_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/kataras/iris/v12/httptest"
	"github.com/monetrapp/rest-api/pkg/application"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/controller"
	"github.com/monetrapp/rest-api/pkg/internal/plaid_helper"
	"github.com/monetrapp/rest-api/pkg/internal/testutils"
	"github.com/plaid/plaid-go/plaid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func NewTestApplication(t *testing.T) *httptest.Expect {
	configuration := config.Configuration{
		Name:          t.Name(),
		UIDomainName:  "http://localhost:1234",
		APIDomainName: "http://localhost:1235",
		AllowSignUp:   true,
		JWT: config.JWT{
			LoginJwtSecret:        gofakeit.UUID(),
			RegistrationJwtSecret: gofakeit.UUID(),
		},
		PostgreSQL: config.PostgreSQL{},
		SMTP:       config.SMTPClient{},
		ReCAPTCHA:  config.ReCAPTCHA{},
		Plaid: config.Plaid{
			Environment: plaid.Sandbox,
		},
		CORS: config.CORS{
			Debug: false,
		},
		Logging: config.Logging{
			Level: "fatal",
		},
	}
	return NewTestApplicationWithConfig(t, configuration)
}

func NewTestApplicationWithConfig(t *testing.T, configuration config.Configuration) *httptest.Expect {
	db := testutils.GetPgDatabase(t)
	p := plaid_helper.NewPlaidClient(logrus.WithField("test", t.Name()), plaid.ClientOptions{
		ClientID:    configuration.Plaid.ClientID,
		Secret:      configuration.Plaid.ClientSecret,
		Environment: configuration.Plaid.Environment,
		HTTPClient:  http.DefaultClient,
	})

	mockJobManager := testutils.NewMockJobManager()

	c := controller.NewController(configuration, db, mockJobManager, p, nil, nil)
	app := application.NewApp(configuration, c)
	return httptest.New(t, app)
}

func GivenIHaveToken(t *testing.T, e *httptest.Expect) string {
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
	token := response.JSON().Path("$.token").String().Raw()
	require.NotEmpty(t, token, "token cannot be empty")

	return token
}
