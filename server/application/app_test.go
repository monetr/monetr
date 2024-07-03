package application_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/application"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/ui"
	"github.com/plaid/plaid-go/v20/plaid"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	log := testutils.GetLog(t)
	conf := config.Configuration{
		AllowSignUp: true,
		Server: config.Server{
			ExternalURL: "http://monetr.local",
			Cookies: config.Cookies{
				SameSiteStrict: true,
				Secure:         true,
				Name:           "M-Token",
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
	assert.NotNil(t, app)
}
