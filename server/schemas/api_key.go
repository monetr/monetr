package schemas

import "github.com/monetr/validation"

type CreateApiKeyRequest struct {
	Name string `json:"name"`
	// Challenge and Nonce are only present when proof of work is enabled.
	Challenge string `json:"challenge"`
	Nonce     uint64 `json:"nonce"`
}

type DeleteApiKeyRequest struct {
	// Challenge and Nonce are only present when proof of work is enabled.
	Challenge string `json:"challenge"`
	Nonce     uint64 `json:"nonce"`
}

var (
	CreateApiKey = validation.Map(
		validation.Key("name",
			Name(),
		).Required(Require),
	)

	// CreateApiKeyChallenge is used instead of CreateApiKey when proof of work is
	// enabled, it additionally requires a solved challenge.
	CreateApiKeyChallenge = validation.Map(
		validation.Key("name",
			Name(),
		).Required(Require),
		validation.Key("challenge",
			Challenge(),
		).Required(Require),
		validation.Key("nonce",
			Nonce(),
		).Required(Require),
	)

	// DeleteApiKeyChallenge validates the body of a delete request when proof of
	// work is enabled. When it is disabled the delete request has no body.
	DeleteApiKeyChallenge = validation.Map(
		validation.Key("challenge",
			Challenge(),
		).Required(Require),
		validation.Key("nonce",
			Nonce(),
		).Required(Require),
	)
)
