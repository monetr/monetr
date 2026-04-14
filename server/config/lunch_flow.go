package config

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

const DefaultLunchFlowAPIURL = "https://lunchflow.app/api/v1"

type LunchFlow struct {
	// Enabled just determines whether or not Lunch Flow will be an option to
	// configure in the UI. This defaults to true as it requires no additional
	// configuration here for self-hosted users.
	Enabled bool `yaml:"enabled"`
	// AllowedApiUrls is the set of Lunch Flow API URLs this deployment is
	// permitted to contact. Comparison is exact-string.
	AllowedApiUrls []string `yaml:"allowedApiUrls"`
}

// ValidateConfig can be called at startup in order to catch problems with the
// configuration early on. If lunch flow is not enabled then this is a no-op, if
// lunch flow is enabled and there are allowed URLs specified; then this will
// validate that those URLs specified are all valid.
func (l LunchFlow) ValidateConfig() error {
	if !l.IsEnabled() {
		return nil
	}
	for _, allowed := range l.AllowedApiUrls {
		parsed, err := url.Parse(allowed)
		if err != nil {
			return errors.Wrapf(err, "configured Lunch Flow url (%s) is not valid", allowed)
		}

		// Do not allow query parameters in the URL as these will be removed when
		// requests are made!
		if len(parsed.Query()) > 0 {
			return errors.Errorf("Lunch Flow url (%s) cannot contain query parameters", allowed)
		}

		// Require a scheme to be specified
		switch strings.ToLower(parsed.Scheme) {
		case "http", "https":
			// These are considered valid!
		default:
			// Any other scheme is not considered valid here!
			return errors.Errorf("Lunch Flow url (%s) must use an http or https scheme", allowed)
		}
	}

	return nil
}

// IsEnabled only returns true if the lunch flow integration is enabled AND when
// there is at least one allowed API URLs configured.
func (l LunchFlow) IsEnabled() bool {
	return l.Enabled && len(l.AllowedApiUrls) > 0
}

// IsAllowedApiUrl returns true when the provided URL matches one of the
// configured allowed Lunch Flow API URLs. Matching is done after both the input
// URL and the allowed URLs are parsed via [url.Parse] in order to ensure
// correctness.
func (l LunchFlow) IsAllowedApiUrl(input string) bool {
	inputUrl, err := url.Parse(input)
	if err != nil {
		return false
	}

	for _, allowed := range l.AllowedApiUrls {
		allowedUrl, err := url.Parse(allowed)
		if err != nil {
			// If an allowed URL in the configuration is not even considered a valid
			// url then discard it. It will not be considered valid in the http client
			// anyway.
			continue
		}
		// Urls must be equal AFTER parsing, the [url.Parse] function does some
		// transformations here that are considered reasonable. Such as converting
		// the scheme to be lowercase.
		if allowedUrl.String() == inputUrl.String() {
			return true
		}
	}
	return false
}
