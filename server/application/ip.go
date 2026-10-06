package application

import (
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/monetr/monetr/server/config"
)

// newIPExtractor builds the strategy echo uses for ctx.RealIP(). Headers are
// only trusted when we know which proxies are allowed to set them, without any
// trusted proxies we use the address of the connection itself. The trusted
// proxies are validated at startup, so if they somehow fail to parse here we
// fall back to the safe option.
func newIPExtractor(server config.Server) echo.IPExtractor {
	trustedProxies, err := server.GetTrustedProxies()
	if err != nil || len(trustedProxies) == 0 {
		return echo.ExtractIPDirect()
	}

	header := server.GetClientIPHeader()
	if header == echo.HeaderXForwardedFor {
		// The trust options that default to true in echo (loopback, link local and
		// private networks) are turned off so that only the configured ranges are
		// skipped, otherwise a client could put a private address in the header and
		// have it treated as a proxy hop.
		options := []echo.TrustOption{
			echo.TrustLoopback(false),
			echo.TrustLinkLocal(false),
			echo.TrustPrivateNet(false),
		}
		for _, trustedProxy := range trustedProxies {
			options = append(options, echo.TrustIPRange(trustedProxy))
		}
		return echo.ExtractIPFromXFFHeader(options...)
	}

	// We don't use echo's ExtractIPFromRealIPHeader for other headers because it
	// only checks whether the address inside the header is trusted, not whether
	// the connection came from a trusted proxy. So we do that ourselves here.
	direct := echo.ExtractIPDirect()
	return func(req *http.Request) string {
		directIP := direct(req)
		if !isTrusted(trustedProxies, net.ParseIP(directIP)) {
			return directIP
		}

		values := req.Header.Values(header)
		if len(values) == 0 {
			return directIP
		}
		// If a client sent the header too and the proxy appended to it instead of
		// replacing it, the proxy's value is the last one.
		items := strings.Split(values[len(values)-1], ",")
		if ip := parseHeaderIP(items[len(items)-1]); ip != nil {
			return ip.String()
		}

		return directIP
	}
}

// newRequestIDHeaderFilter strips the request ID headers off of any request
// that did not come directly from a trusted proxy. Same as the client IP
// header, a client could otherwise set these to whatever they want and have it
// show up in our logs and traces.
func newRequestIDHeaderFilter(server config.Server) echo.MiddlewareFunc {
	trustedProxies, err := server.GetTrustedProxies()
	if err != nil {
		trustedProxies = nil
	}
	direct := echo.ExtractIPDirect()
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx *echo.Context) error {
			req := ctx.Request()
			if !isTrusted(trustedProxies, net.ParseIP(direct(req))) {
				req.Header.Del("X-Request-Id")
				req.Header.Del("X-Cloud-Trace-Context")
			}
			return next(ctx)
		}
	}
}

func isTrusted(trustedProxies []*net.IPNet, ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, trustedProxy := range trustedProxies {
		if trustedProxy.Contains(ip) {
			return true
		}
	}
	return false
}

// parseHeaderIP parses an address out of a header value, allowing for IPv6
// brackets and a port since some proxies include one.
func parseHeaderIP(value string) net.IP {
	value = strings.TrimSpace(value)
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	}
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	return net.ParseIP(value)
}
