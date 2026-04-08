package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type Cookies struct {
	// SameSiteStrict allows the host of monetr to define whether the cookie used
	// for authentication is limited to same site. This might impact use cases
	// where the UI is on a different domain than the API. In general, it is
	// recommended that this is enabled and that the UI and API are served from
	// the same domain.
	SameSiteStrict bool `yaml:"sameSiteStrict"`
	// Secure specifies that the authentication cookie issued and required by API
	// endpoints is a secure cookie. This defaults to true, but requires that the
	// host of monetr use HTTPS. If you are not using HTTPS then this must be
	// disabled for API calls to succeed.
	Secure bool `yaml:"secure"`
	// Determines whether or not the cookies issued to clients will be HTTP only
	// cookies. HTTP only cookies cannot be read by the javascript executed in
	// client browsers, as such they are slightly more secure. This setting
	// defaults to true and should only be disabled if you understand what you are
	// doing and need access to the token in a client side application.
	HttpOnly bool `yaml:"httpOnly"`
	// Name defines the name of the cookie to use for authentication. This
	// defaults to `M-Token` but can be customized if the host wants to.
	Name string `yaml:"name"`
}

type Server struct {
	// ListenPort defines the port that monetr will listen for HTTP requests on.
	// This port should be forwarded such that it is accessible to the desired
	// clients. Be that on a local network, or forwarded to the public internet.
	ListenPort int `yaml:"listenPort"`
	// ListenAddress defines the IP address that monetr should listen on for HTTP
	// requests.
	ListenAddress string `yaml:"listenAddress"`
	// StatsPort is the port that our prometheus metrics are served on. This port
	// should not be publicly accessible and should only be accessible by the
	// prometheus server scraping for metrics. It is not an endpoint that needs to
	// be secured as no sensitive client information will be served by it; but it
	// should not be accessible publicly.
	StatsPort int `yaml:"statsPort"`
	// Cookies defines the parameters used for issuing and processing cookies from
	// clients. Cookies are used for authentication.
	Cookies Cookies `yaml:"cookies"`
	// UICacheHours is the number of hours that UI files should be cached by the
	// client. This is done by including an Expires and Cache-Control header in
	// the response for all UI related requests. If this is 0 then the headers
	// will not be included. Defaults to 14 days (336 hours).
	UICacheHours int `yaml:"uiCacheHours"`
	// ExternalURL tells monetr what protocol, hostname and path it should expect
	// traffic from externally. For example: `http://my.monetr.local` tells monetr
	// that it should not expect secure traffic, and thus should not use things
	// like secure cookies (as they will not work). Where as
	// `https://my.monetr.local` tells monetr that all traffic will be via HTTPS
	// and secure cookies will work. Another example would be something like
	// `https://homelab.local/monetr` where monetr is on the same domain as
	// potentially other applications, but is under a specific sub path.
	ExternalURL string `yaml:"externalUrl"`
	// TLS Certificate for the API server listener. This also requires TLS Key.
	// This can be the certificate content, or a filepath to the certificate on
	// the filesystem. If it is a filepath then monetr must have sufficient
	// permission to access the certificate.
	// At the moment, this certificate will not automatically rotate. If the
	// certificate changes on the filesystem, then the server needs to be
	// restarted.
	TLSCertificate string `yaml:"tlsCertificate"`
	// TLS Key for the API server listener. This also requires TLS Certificate.
	// This can be the key content, or a filepath to the key on the filesystem. If
	// it is a filepath then monetr must have sufficient permimssion to access the
	// key.
	// At the moment, this certificate will not automatically rotate. If the
	// certificate changes on the filesystem, then the server needs to be
	// restarted.
	TLSKey string `yaml:"tlsKey"`
}

// GetIsSecureProtocol will return true if the ExternalURL specified is a secure
// url using HTTPS.
func (s Server) GetIsSecureProtocol() bool {
	return strings.HasPrefix(s.ExternalURL, "https://")
}

// GetIsCookieSecure will return true when both Cookies.Secure is true and the
// ExternalURL has been configured to use HTTPS. This is used to determine
// whether or not to set `secure` on the authentication cookies monetr issues to
// clients.
func (s Server) GetIsCookieSecure() bool {
	return s.Cookies.Secure && s.GetIsSecureProtocol()
}

// AssertExternalURLValid will return an error if the specified ExternalURL is
// not valid.
func (s Server) AssertExternalURLValid() error {
	_, err := url.Parse(s.ExternalURL)
	if err != nil {
		return errors.Wrap(err, "external URL is not valid")
	}

	return nil
}

// GetHostname will return the hostname derived from the ExternalURL, it will
// not include a port if one was specified. This is used for setting cookies.
func (s Server) GetHostname() string {
	url := s.GetBaseURL()
	return url.Hostname()
}

func (s Server) GetBaseURL() *url.URL {
	url, err := url.Parse(s.ExternalURL)
	if err != nil {
		address := s.GetListenAddress()
		url, _ = url.Parse(fmt.Sprintf("http://%s:%d", address, s.ListenPort))
	}

	return url
}

// GetListenAddress returns a valid IP address that will be used to determine
// the listen address for the server. Or localhost if the address is not valid
// or not provided.
func (s Server) GetListenAddress() string {
	if s.ListenAddress == "" {
		return "localhost"
	}

	addr := net.ParseIP(s.ListenAddress)
	if addr == nil {
		return "localhost"
	}

	return addr.String()
}

// GetURL should be used to generate URLs with paths and parameters that are
// safe to use outside the application. For example, if you are sending an email
// with a link to a page in monetr, this function should be used to generate
// that URL.
func (s Server) GetURL(relativePath string, params map[string]string) string {
	baseUrl := s.GetBaseURL().JoinPath(relativePath)
	query := baseUrl.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	baseUrl.RawQuery = query.Encode()

	return baseUrl.String()
}
