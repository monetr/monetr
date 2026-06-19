package testutils

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/config"
	"github.com/plaid/plaid-go/v42/plaid"
)

const (
	TestEmailDomain = "monetr.local"
)

func GetConfig(_ *testing.T) config.Configuration {
	return config.Configuration{
		AllowSignUp: true,
		Features: config.Features{
			// Enabled for test suites and local dev
			TransactionImports: true,
		},
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
		ProofOfWork: config.ProofOfWork{
			// Proof of work is disabled by default in tests so that the existing
			// test suite does not need to solve a challenge on every auth request.
			// Tests that specifically exercise proof of work flip Enabled on. The
			// difficulty is kept low (8) so that when it IS enabled the tests stay
			// fast and do not flake on the larger variance of higher difficulties.
			Enabled:    false,
			Difficulty: 8,
			Lifetime:   5 * time.Minute,
		},
		LunchFlow: config.LunchFlow{
			// By default lunch flow is enabled in tests, disable it to simulate
			// alternate behaviors.
			Enabled:        true,
			AllowedApiUrls: []string{"https://www.lunchflow.app/api/v1"},
		},
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
