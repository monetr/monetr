package application_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/application"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/ui"
	"github.com/plaid/plaid-go/v41/plaid"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	log := testutils.GetLog(t)
	conf := config.Configuration{
		AllowSignUp: true,
		Server: config.Server{
			ExternalURL: "https://monetr.local",
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

func TestNewAppServerTimeouts(t *testing.T) {
	app := application.NewApp(config.Configuration{})

	assert.Equal(t, 5*time.Second, app.Server.ReadHeaderTimeout, "ReadHeaderTimeout should be set to mitigate slowloris")
	assert.Equal(t, 30*time.Second, app.Server.ReadTimeout, "ReadTimeout should be set to mitigate slowloris")
	assert.Equal(t, 45*time.Second, app.Server.WriteTimeout, "WriteTimeout must exceed the 30s Plaid long-poll ceiling in controller.getWaitForPlaid")
	assert.Equal(t, 120*time.Second, app.Server.IdleTimeout, "IdleTimeout should bound keep-alive idle duration")
}
