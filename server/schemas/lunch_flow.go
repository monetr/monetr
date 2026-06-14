package schemas

import (
	"net/url"
	"strings"

	"github.com/monetr/monetr/server/validators"
	"github.com/monetr/validation"
	"github.com/monetr/validation/is"
)

type PostLunchFlowLinkRequest struct {
	Name         string `json:"name"`
	LunchFlowURL string `json:"lunchFlowURL"`
	APIKey       string `json:"apiKey"`
}

var (
	// PostLunchFlowLink is intended to be used alongside
	// [PostLunchFlowLinkRequest] in order to parse a non model request body.
	PostLunchFlowLink = validation.Map(
		Name(Require),
		validation.Key(
			"lunchFlowURL",
			validation.Required.Error("Lunch Flow API URL is required to setup a Lunch Flow link"),
			validation.IsString,
			LunchFlowAPIURL(),
		).Required(Require),
		validation.Key(
			"apiKey",
			validation.Required.Error("Lunch Flow API Key must be provided to setup a Lunch Flow link"),
			validation.IsString,
			validation.Length(1, 100).Error("Lunch Flow API Key must be between 1 and 100 characters"),
			is.UTFLetterNumeric,
		).Required(validators.Require),
	)
)

func LunchFlowAPIURL() validation.Rule {
	return validation.NewStringRule(func(input string) bool {
		parsed, err := url.Parse(input)
		if err != nil {
			return false
		}
		// Do not allow query parameters in the URL as these will be removed
		// when requests are made!
		if len(parsed.Query()) > 0 {
			return false
		}

		// Require a scheme to be specified
		switch strings.ToLower(parsed.Scheme) {
		case "http", "https":
			// These are considered valid!
		default:
			// Any other scheme is not considered valid here!
			return false
		}

		// If the URL has any credentials in the actual URL then something goofy
		// is going on and it should be rejected.
		if parsed.User != nil {
			if parsed.User.Username() != "" {
				return false
			}

			if pass, _ := parsed.User.Password(); pass != "" {
				return false
			}
		}

		return true
	}, "Lunch Flow API URL must be a full valid URL")
}
