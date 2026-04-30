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
			cspPolicy, trustedTypesPolicy = c.buildPolicies()
		},
	)

	if c.configuration.Sentry.SecurityHeaderEndpoint != "" {
		ctx.Response().Header().Set(
			"Report-To",
			// TODO properly json encode this before hand, atm the security header
			// endpoint is not properly escaped.
			fmt.Sprintf(
				`{"group":"csp-endpoint","max_age":1800,"endpoints":[{"url":%q}],"include_subdomains":true}`,
				c.configuration.Sentry.SecurityHeaderEndpoint,
			),
		)
		// Reporting-Endpoints is the standardized successor to Report-To. The
		// Report-To header above is kept around as a fallback, but at the time of
		// writing this any browser that ships Reporting-Endpoints prefers it over
		// the legacy header, so Reporting-Endpoints is the path monetr expects most
		// CSP reports to be delivered through. The named "csp-endpoint" group here
		// matches the "report-to csp-endpoint" directive embedded in the CSP value
		// below, so the same endpoint is referenced regardless of which version of
		// the Reporting API the user agent supports. The structured-fields encoding
		// (RFC 8941) requires the URL to be a quoted string with " and \ backslash
		// escaped, which is exactly what fmt's %q verb produces; URLs per RFC 3986
		// only contain ASCII printable characters so the output is always SF
		// conformant. See
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Reporting-Endpoints
		ctx.Response().Header().Set(
			"Reporting-Endpoints",
			fmt.Sprintf("csp-endpoint=%q", c.configuration.Sentry.SecurityHeaderEndpoint),
		)

		// Integrity-Policy-Report-Only asks the browser to flag any <script> or
		// <link rel="stylesheet"> fetch that is missing inline integrity metadata
		// (the integrity="sha512-..." attribute). The first-party bundle already
		// emits SRI for every static and dynamic chunk via rsbuild's SRI plugin,
		// but third-party scripts injected at runtime (Plaid Link, ReCAPTCHA, and
		// possibly Sentry's lazy integrations) do not carry integrity attributes
		// and would be blocked under the enforcing variant. Sending this as
		// report-only lets monetr collect violation reports through the same
		// csp-endpoint without breaking those integrations, so the data can guide a
		// future move to the enforcing Integrity-Policy header. See
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Integrity-Policy
		// The value is a structured-field dictionary (RFC 8941); inner lists use
		// space separators inside the parens and dictionary keys are separated by
		// commas. Only set when a security header endpoint is configured because
		// without an endpoint there is nowhere for the reports to go.
		ctx.Response().Header().Set(
			"Integrity-Policy-Report-Only",
			"blocked-destinations=(script), endpoints=(csp-endpoint)",
		)
	}

	ctx.Response().Header().Set("Content-Security-Policy", cspPolicy)
	ctx.Response().Header().Set("Content-Security-Policy-Report-Only", trustedTypesPolicy)
}

// buildPolicies assembles the Content-Security-Policy and the report-only
// Trusted Types policy strings based on the active configuration. It is split
// out of ApplyContentSecurityPolicy so the assembly can be exercised by tests
// without having to reset the package level sync.Once cache. The returned
// strings depend only on configuration, which does not change at runtime, so
// caching them once per process via cspPolicyFunc remains safe.
func (c *UIController) buildPolicies() (csp string, trustedTypes string) {
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

	csp = encodePolicies()

	// Trusted Types asks the browser to require that strings passed to dangerous
	// sinks like innerHTML be wrapped in a TrustedHTML object that was created by
	// a named policy. This helps protect against DOM based cross-site scripting
	// attacks. See
	// https://developer.mozilla.org/en-US/docs/Web/API/Trusted_Types_API
	// At the time of writing this, we cannot be sure that every dependency the UI
	// relies on (Sentry, Plaid, ReCAPTCHA, and so on) is compatible with Trusted
	// Types, so this is sent as a Content-Security-Policy-Report-Only header.
	// Violations are sent to the same endpoint as the rest of the CSP reports
	// without actually breaking the page. Once the reports are quiet this can be
	// promoted into the enforcing Content-Security-Policy header.
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
	trustedTypes = strings.Join(trustedTypesParts, "; ")

	return csp, trustedTypes
}
