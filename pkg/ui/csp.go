package ui

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	ua "github.com/mileusna/useragent"
)

const (
	Self         = "'self'"
	UnsafeInline = "'unsafe-inline'"
)

var (
	noop = struct{}{}
)

// At the time of writing this, it seems that Chrome is the only browser engine that has implemented a majority of the
// content security policy items. See https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
// Because of this, it is not easy (or reasonable) to have a single CSP header that will securely cover every browser.
// This code aims to provide a header based on the user-agent of the browser in the request.

func (c *UIController) ContentSecurityPolicyMiddleware(ctx echo.Context) {
	userAgentString := ctx.Request().UserAgent()
	userAgent := ua.Parse(userAgentString)

	policies := map[string]map[string]struct{}{
		"default-src": {
			Self:                    noop,
			"https://cdn.plaid.com": noop,
		},
		"script-src-elem": {
			Self:                  noop,
			"https://*.plaid.com": noop,
			UnsafeInline:          noop,
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
			Self:                  noop,
			"https://*.plaid.com": noop,
		},
		"img-src": {
			Self:    noop,
			"data:": noop,
		},
	}

	// Only allow google to connect when ReCAPTCHA is enabled.
	if c.configuration.ReCAPTCHA.Enabled {
		policies["script-src-elem"]["https://www.gstatic.com"] = noop
		policies["script-src-elem"]["https://www.google.com"] = noop
		policies["frame-src"]["https://www.google.com"] = noop
	}

	// If sentry is enabled and a DSN is configured, then setup the connect-src for sentry.
	if c.configuration.Sentry.Enabled {
		policies["connect-src"]["https://sentry.io"] = noop
		if c.configuration.Sentry.ExternalDSN != "" {
			if dsn, err := url.Parse(c.configuration.Sentry.ExternalDSN); err == nil {
				policies["connect-src"][fmt.Sprintf("https://%s", dsn.Hostname())] = noop
				policies["script-src-elem"][fmt.Sprintf("https://%s", dsn.Hostname())] = noop
			}
		} else if c.configuration.Sentry.DSN != "" {
			if dsn, err := url.Parse(c.configuration.Sentry.DSN); err == nil {
				policies["connect-src"][fmt.Sprintf("https://%s", dsn.Hostname())] = noop
				policies["script-src-elem"][fmt.Sprintf("https://%s", dsn.Hostname())] = noop
			}
		}
	}

	encodePolicies := func() string {
		policyParts := make([]string, 0, len(policies))

		for kind, items := range policies {
			part := make([]string, 0, len(items)+1)
			part = append(part, kind)
			for item := range items {
				part = append(part, item)
			}
			policyParts = append(policyParts, strings.Join(part, " "))
		}

		return strings.Join(policyParts, "; ")
	}

	switch {
	case userAgent.IsChrome() || (!userAgent.IsIOS() && userAgent.IsSafari()):
		// Safari and Chrome desktop seem to work.
	case userAgent.IsFirefox() || (userAgent.IsIOS() && userAgent.IsSafari()):
		{ // script-src-elem is not supported on firefox, or safari for ios.
			for item := range policies["script-src-elem"] {
				policies["default-src"][item] = noop
			}
			delete(policies, "script-src-elem")
		}

		{ // style-src-elem is not supported on firefox, or safari for ios.
			for item := range policies["style-src-elem"] {
				policies["default-src"][item] = noop
			}
			delete(policies, "style-src-elem")
		}
	case userAgent.IsInternetExplorer():
		// No CSP policies for IE. If you're using it you hate security anyway.
		policies = map[string]map[string]struct{}{}
	default:
		{ // script-src-elem is not supported on firefox, or safari for ios.
			for item := range policies["script-src-elem"] {
				policies["default-src"][item] = noop
			}
			delete(policies, "script-src-elem")
		}

		{ // style-src-elem is not supported on firefox, or safari for ios.
			for item := range policies["style-src-elem"] {
				policies["default-src"][item] = noop
			}
			delete(policies, "style-src-elem")
		}
	}

	if len(policies) > 0 {
		ctx.Response().Header().Set("Content-Security-Policy", encodePolicies())
	}
}
