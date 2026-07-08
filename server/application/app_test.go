package application_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/application"
	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/ui"
	"github.com/plaid/plaid-go/v43/plaid"
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

type testController struct{}

func (testController) RegisterRoutes(app *echo.Echo) {
	app.GET("/test", func(ctx *echo.Context) error {
		return ctx.NoContent(http.StatusOK)
	})
}

func newTestApplication(t *testing.T, configuration config.Configuration) *httpexpect.Expect {
	log := testutils.GetLog(t)
	app := application.NewApp(configuration, testController{})

	// run server using httptest
	server := httptest.NewServer(app)
	t.Cleanup(func() {
		server.Close()
	})

	return httpexpect.WithConfig(httpexpect.Config{
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
		Context: t.Context(),
	})
}

func TestNewAppHSTSHeader(t *testing.T) {
	t.Run("sets HSTS for HTTPS external URL", func(t *testing.T) {
		e := newTestApplication(t, config.Configuration{
			Server: config.Server{
				ExternalURL: "https://monetr.local",
			},
		})

		response := e.GET("/test").Expect()
		response.Status(http.StatusOK)
		response.Header("Strict-Transport-Security").IsEqual("max-age=31536000")
	})

	t.Run("omits HSTS for HTTP external URL", func(t *testing.T) {
		e := newTestApplication(t, config.Configuration{
			Server: config.Server{
				ExternalURL: "http://monetr.local",
			},
		})

		response := e.GET("/test").Expect()
		response.Status(http.StatusOK)
		response.Headers().NotContainsKey("Strict-Transport-Security")
	})
}

func TestNewAppServerTimeouts(t *testing.T) {
	// echo v5 no longer exposes the http.Server, the timeouts now get applied via
	// ConfigureServer in StartConfig.BeforeServeFunc so we assert against that.
	server := &http.Server{}
	application.ConfigureServer(server)

	assert.Equal(t, 5*time.Second, server.ReadHeaderTimeout, "ReadHeaderTimeout should be set to mitigate slowloris")
	assert.Equal(t, 30*time.Second, server.ReadTimeout, "ReadTimeout should be set to mitigate slowloris")
	assert.Equal(t, 45*time.Second, server.WriteTimeout, "WriteTimeout must exceed the 30s Plaid long-poll ceiling in controller.getWaitForPlaid")
	assert.Equal(t, 120*time.Second, server.IdleTimeout, "IdleTimeout should bound keep-alive idle duration")
}
