package config

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const DefaultSimpleFinHost = "bridge.simplefin.org"

type SimpleFin struct {
	// Enabled determines whether or not SimpleFIN will be an option to configure
	// in the UI. This defaults to true as it requires no additional configuration
	// for self-hosted users.
	Enabled bool `yaml:"enabled"`
	// AllowedHosts is the set of SimpleFIN hosts this deployment is permitted to
	// contact. This is a host and not a full URL like Lunch Flow because the
	// SimpleFIN setup token has the claim token in the URL's path. Hosts should
	// be specified like `bridge.simplefin.org`, with a port only if the server is
	// on a non-standard one.
	AllowedHosts []string `yaml:"allowedHosts"`
}

// ValidateConfig can be called at startup in order to catch problems with the
// configuration early on. If SimpleFIN is not enabled then this is a no-op.
// Otherwise every allowed host must be just a host and optional port.
func (s SimpleFin) ValidateConfig() error {
	if !s.IsEnabled() {
		return nil
	}
	for _, allowed := range s.AllowedHosts {
		// Parse it as if it were the host of an https URL, if anything other than
		// the host is present in the string then it will show up somewhere else on
		// the parsed URL and the host won't match.
		parsed, err := url.Parse("https://" + allowed)
		if err != nil {
			return errors.Wrapf(err, "configured SimpleFIN host (%s) is not valid", allowed)
		}

		if parsed.Host != allowed || parsed.Hostname() == "" {
			return errors.Errorf("SimpleFIN host (%s) must only be a host and an optional port", allowed)
		}
	}

	return nil
}

// IsEnabled only returns true if the SimpleFIN integration is enabled AND when
// there is at least one allowed host configured.
func (s SimpleFin) IsEnabled() bool {
	return s.Enabled && len(s.AllowedHosts) > 0
}

// IsAllowedUrl returns true when the provided URL is an https URL whose host
// matches one of the configured allowed SimpleFIN hosts. The host comparison is
// case insensitive and includes the port, so `bridge.simplefin.org:443` will
// not match `bridge.simplefin.org`.
func (s SimpleFin) IsAllowedUrl(input string) bool {
	inputUrl, err := url.Parse(input)
	if err != nil {
		return false
	}

	// SimpleFIN sends credentials with every request, so plain http is never
	// allowed.
	if !strings.EqualFold(inputUrl.Scheme, "https") {
		return false
	}

	for _, allowed := range s.AllowedHosts {
		if strings.EqualFold(inputUrl.Host, allowed) {
			return true
		}
	}

	return false
}
