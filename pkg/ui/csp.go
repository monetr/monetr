package ui

import (
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/mileusna/useragent"
)

const (
	Self         = "'self'"
	UnsafeInline = "'unsafe-inline'"
)

var (
	noop = struct{}{}
)

var (
	_ iris.Handler = (&UIController{}).ContentSecurityPolicyMiddleware
)

// At the time of writing this, it seems that Chrome is the only browser engine that has implemented a majority of the
// content security policy items. See https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
// Because of this, it is not easy (or reasonable) to have a single CSP header that will securely cover every browser.
// This code aims to provide a header based on the user-agent of the browser in the request.

func (c *UIController) ContentSecurityPolicyMiddleware(ctx iris.Context) {
	userAgentString := ctx.GetHeader("User-Agent")
	userAgent := ua.Parse(userAgentString)

	policies := map[string]map[string]struct{}{
		"default-src": {
			Self: noop,
		},
		"script-src-elem": {
			Self:                      noop,
			"https://www.gstatic.com": noop,
			"https://www.google.com":  noop,
			"https://*.plaid.com":     noop,
		},
		"font-src": {
			Self: noop,
		},
		"style-src-elem": {
			UnsafeInline: noop,
		},
		"connect-src": {
			Self: noop, // Add ws if its in development mode.
		},
		"frame-src": {
			Self:                     noop,
			"https://*.plaid.com":    noop,
			"https://www.google.com": noop, // For ReCAPTCHA
		},
		"img-src": {
			Self:    noop,
			"data:": noop,
		},
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
	case userAgent.IsChrome():
		// Chrome actually works.
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
	}

	if len(policies) > 0 {
		ctx.Header("Content-Security-Policy", encodePolicies())
	}

	ctx.Next()
}
