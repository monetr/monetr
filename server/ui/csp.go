package ui

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

const (
	Self         = "'self'"
	UnsafeInline = "'unsafe-inline'"
)

var (
	noop                      = struct{}{}
	cspPolicy          string = ""
	trustedTypesPolicy string = ""
	cspPolicyFunc      sync.Once
)

// At the time of writing this, it seems that Chrome is the only browser engine
// that has implemented a majority of the content security policy items. See
// https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP Because of this, it is
// not easy (or reasonable) to have a single CSP header that will securely cover
// every browser. This code aims to provide a header based on the user-agent of
// the browser in the request.
func (c *UIController) ApplyContentSecurityPolicy(ctx echo.Context) {
	cspPolicyFunc.Do(
		func() {
			policies := map[string]map[string]struct{}{
				"default-src": {
					Self: noop,
				},
				"script-src-elem": {
					Self:         noop,
					UnsafeInline: noop,
				},
				"font-src": {
					Self:    noop,
					"data:": noop,
				},
				"style-src-elem": {
					UnsafeInline: noop,
					Self:         noop,
				},
				"connect-src": {
					Self: noop, // Add ws if its in development mode.
				},
				"frame-src": {
					Self: noop,
				},
				"img-src": {
					Self:    noop,
					"data:": noop,
				},
				"report-uri": {},
				"report-to":  {},
			}

			if c.configuration.Plaid.GetEnabled() {
				policies["default-src"]["https://cdn.plaid.com"] = noop
				policies["script-src-elem"]["https://*.plaid.com"] = noop
				policies["frame-src"]["https://*.plaid.com"] = noop
			}

			// Only allow google to connect when ReCAPTCHA is enabled.
			if c.configuration.ReCAPTCHA.Enabled {
				policies["script-src-elem"]["https://www.gstatic.com"] = noop
				policies["script-src-elem"]["https://www.google.com"] = noop
				policies["frame-src"]["https://www.google.com"] = noop
			}

			// If sentry is enabled and an external DSN is configured, then setup the
			// connect-src for sentry.
			if c.configuration.Sentry.Enabled {
				if c.configuration.Sentry.ExternalDSN != "" {
					policies["connect-src"]["https://sentry.io"] = noop
					if dsn, err := url.Parse(c.configuration.Sentry.ExternalDSN); err == nil {
						policies["connect-src"][fmt.Sprintf("%s://%s", dsn.Scheme, dsn.Hostname())] = noop
						policies["script-src-elem"][fmt.Sprintf("%s://%s", dsn.Scheme, dsn.Hostname())] = noop
					}
				}

				if c.configuration.Sentry.SecurityHeaderEndpoint != "" {
					policies["report-uri"][c.configuration.Sentry.SecurityHeaderEndpoint] = noop
					policies["report-to"]["csp-endpoint"] = noop
				}
			}

			encodePolicies := func() string {
				policyParts := make([]string, 0, len(policies))

				for kind, items := range policies {
					if len(items) == 0 {
						continue
					}

					part := make([]string, 0, len(items)+1)
					part = append(part, kind)
					for item := range items {
						part = append(part, item)
					}
					policyParts = append(policyParts, strings.Join(part, " "))
				}

				return strings.Join(policyParts, "; ")
			}

			cspPolicy = encodePolicies()

			// Trusted Types asks the browser to require that strings passed to
			// dangerous sinks like innerHTML be wrapped in a TrustedHTML object that
			// was created by a named policy. This helps protect against DOM based
			// cross-site scripting attacks. See
			// https://developer.mozilla.org/en-US/docs/Web/API/Trusted_Types_API
			// At the time of writing this, we cannot be sure that every dependency
			// the UI relies on (Sentry, Plaid, ReCAPTCHA, and so on) is compatible
			// with Trusted Types, so this is sent as a
			// Content-Security-Policy-Report-Only header. Violations are sent to the
			// same endpoint as the rest of the CSP reports without actually breaking
			// the page. Once the reports are quiet this can be promoted into the
			// enforcing Content-Security-Policy header.
			trustedTypesParts := []string{
				"require-trusted-types-for 'script'",
				"trusted-types react",
			}
			if c.configuration.Sentry.SecurityHeaderEndpoint != "" {
				trustedTypesParts = append(
					trustedTypesParts,
					"report-uri "+c.configuration.Sentry.SecurityHeaderEndpoint,
					"report-to csp-endpoint",
				)
			}
			trustedTypesPolicy = strings.Join(trustedTypesParts, "; ")
		},
	)

	if c.configuration.Sentry.SecurityHeaderEndpoint != "" {
		ctx.Response().Header().Set(
			"Report-To",
			// TODO properly json encode this before hand, atm the security header
			// endpoint is not properly escaped.
			fmt.Sprintf(`{"group":"csp-endpoint","max_age":1800,"endpoints":[{"url":%q}],"include_subdomains":true}`, c.configuration.Sentry.SecurityHeaderEndpoint),
		)
	}

	ctx.Response().Header().Set("Content-Security-Policy", cspPolicy)
	ctx.Response().Header().Set("Content-Security-Policy-Report-Only", trustedTypesPolicy)
}
