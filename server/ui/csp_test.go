package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/config"
	"github.com/stretchr/testify/assert"
)

const testCSPEndpoint = "https://o12345.ingest.sentry.io/api/4501234567890/security/?sentry_key=abc123"

// resetCSPCache wipes the package level CSP and Trusted Types cache so a
// subsequent ApplyContentSecurityPolicy call rebuilds the strings from the
// controller's current configuration. The cache is intentionally process-wide
// in production, but tests need to drive multiple configurations in sequence.
func resetCSPCache() {
	cspPolicy = ""
	trustedTypesPolicy = ""
	cspPolicyFunc = sync.Once{}
}

func TestBuildPolicies(t *testing.T) {
	t.Run("sentry csp endpoint configured wires up reporting directives", func(t *testing.T) {
		controller := &UIController{
			configuration: config.Configuration{
				Sentry: config.Sentry{
					Enabled:                true,
					SecurityHeaderEndpoint: testCSPEndpoint,
				},
			},
		}

		csp, trustedTypes := controller.buildPolicies()

		assert.Contains(t, csp, "report-uri "+testCSPEndpoint)
		assert.Contains(t, csp, "report-to csp-endpoint")
		assert.Contains(t, trustedTypes, "require-trusted-types-for 'script'")
		assert.Contains(t, trustedTypes, "trusted-types react")
		assert.Contains(t, trustedTypes, "report-uri "+testCSPEndpoint)
		assert.Contains(t, trustedTypes, "report-to csp-endpoint")
	})

	t.Run("sentry enabled but no csp endpoint omits reporting directives", func(t *testing.T) {
		controller := &UIController{
			configuration: config.Configuration{
				Sentry: config.Sentry{
					Enabled: true,
				},
			},
		}

		csp, trustedTypes := controller.buildPolicies()

		assert.NotContains(t, csp, "report-uri")
		assert.NotContains(t, csp, "report-to")
		assert.Contains(t, trustedTypes, "require-trusted-types-for 'script'")
		assert.Contains(t, trustedTypes, "trusted-types react")
		assert.NotContains(t, trustedTypes, "report-uri")
		assert.NotContains(t, trustedTypes, "report-to")
	})
}

func TestApplyContentSecurityPolicy(t *testing.T) {
	t.Run("sentry csp endpoint configured emits all reporting headers", func(t *testing.T) {
		resetCSPCache()

		controller := &UIController{
			configuration: config.Configuration{
				Sentry: config.Sentry{
					Enabled:                true,
					SecurityHeaderEndpoint: testCSPEndpoint,
				},
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, rec)

		controller.ApplyContentSecurityPolicy(ctx)

		headers := rec.Header()
		assert.Equal(
			t,
			fmt.Sprintf("csp-endpoint=%q", testCSPEndpoint),
			headers.Get("Reporting-Endpoints"),
		)
		assert.Equal(
			t,
			`{"group":"csp-endpoint","max_age":1800,"endpoints":[{"url":"`+testCSPEndpoint+`"}],"include_subdomains":true}`,
			headers.Get("Report-To"),
		)

		csp := headers.Get("Content-Security-Policy")
		assert.Contains(t, csp, "report-uri "+testCSPEndpoint)
		assert.Contains(t, csp, "report-to csp-endpoint")

		trustedTypes := headers.Get("Content-Security-Policy-Report-Only")
		assert.Contains(t, trustedTypes, "require-trusted-types-for 'script'")
		assert.Contains(t, trustedTypes, "trusted-types react")
		assert.Contains(t, trustedTypes, "report-uri "+testCSPEndpoint)
		assert.Contains(t, trustedTypes, "report-to csp-endpoint")
	})

	t.Run("no sentry csp endpoint omits reporting headers", func(t *testing.T) {
		resetCSPCache()

		controller := &UIController{
			configuration: config.Configuration{
				Sentry: config.Sentry{
					Enabled: true,
				},
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := echo.New().NewContext(req, rec)

		controller.ApplyContentSecurityPolicy(ctx)

		headers := rec.Header()
		assert.Empty(t, headers.Get("Reporting-Endpoints"))
		assert.Empty(t, headers.Get("Report-To"))
		assert.NotEmpty(t, headers.Get("Content-Security-Policy"))
		assert.NotEmpty(t, headers.Get("Content-Security-Policy-Report-Only"))
		assert.NotContains(t, headers.Get("Content-Security-Policy"), "report-uri")
		assert.NotContains(t, headers.Get("Content-Security-Policy"), "report-to")
		assert.NotContains(t, headers.Get("Content-Security-Policy-Report-Only"), "report-uri")
		assert.NotContains(t, headers.Get("Content-Security-Policy-Report-Only"), "report-to")
	})
}
