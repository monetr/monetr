package controller_test

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/application"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/config"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/controller"
	"github.com/harderthanitneedstobe/rest-api/v0/pkg/testutils"
	"github.com/kataras/iris/v12/httptest"
	"github.com/plaid/plaid-go/plaid"
	"testing"
)

func NewTestApplication(t *testing.T) *httptest.Expect {
	db := testutils.GetPgDatabase(t)
	configuration := config.Configuration{
		Name:           t.Name(),
		UIDomainName:   "http://localhost:1234",
		APIDomainName:  "http://localhost:1235",
		AllowSignUp:    true,
		EnableWebhooks: false,
		JWT: config.JWT{
			LoginJwtSecret:        gofakeit.UUID(),
			RegistrationJwtSecret: gofakeit.UUID(),
		},
		PostgreSQL: config.PostgreSQL{},
		SMTP:       config.SMTPClient{},
		ReCAPTCHA:  config.ReCAPTCHA{},
		Plaid: config.Plaid{
			Environment: plaid.Development,
		},
		CORS: config.CORS{
			Debug: false,
		},
	}
	c := controller.NewController(configuration, db)
	app := application.NewApp(configuration, c)
	return httptest.New(t, app)
}
