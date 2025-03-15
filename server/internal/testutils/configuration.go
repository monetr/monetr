package testutils

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/config"
	"github.com/plaid/plaid-go/v30/plaid"
)

const (
	TestEmailDomain = "monetr.local"
)

func GetConfig(t *testing.T) config.Configuration {
	return config.Configuration{
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
		Storage: config.Storage{
			Enabled:  true,
			Provider: "mock",
		},
	}
}
