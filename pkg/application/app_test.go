package application_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/application"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/monetr/monetr/pkg/ui"
	"github.com/plaid/plaid-go/v14/plaid"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	log := testutils.GetLog(t)
	conf := config.Configuration{
		UIDomainName:  "monetr.local",
		APIDomainName: "monetr.local",
		AllowSignUp:   true,
		Server: config.Server{
			Cookies: config.Cookies{
				SameSiteStrict: true,
				Secure:         true,
				Name:           "M-Token",
			},
		},
		JWT: config.JWT{
			LoginJwtSecret:        gofakeit.UUID(),
			RegistrationJwtSecret: gofakeit.UUID(),
			LoginExpiration:       1,
		},
		PostgreSQL: config.PostgreSQL{},
		Email: config.Email{
			Enabled: false,
			Verification: config.EmailVerification{
				Enabled:       false,
				TokenLifetime: 10 * time.Second,
				TokenSecret:   gofakeit.Generate("????????????????"),
			},
			Domain: "monetr.local",
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
	app := application.NewApp(conf, ui.NewUIController(log, conf))
	assert.NotEmpty(t, app.Routes(), "must have some routes registered")
}
