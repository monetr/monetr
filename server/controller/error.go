package controller

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type GenericAPIError interface {
	Cause() error
	Error() string
	FriendlyMessage() string
}

var (
	ErrMFARequired            = errors.New("login requires MFA")
	ErrEmailNotVerified       = errors.New("email address is not verified")
	ErrEmailAlreadyExists     = errors.New("email already in use")
	ErrPasswordChangeRequired = errors.New("password must be changed")
)

var (
	_ GenericAPIError = MFARequiredError{}
	_ json.Marshaler  = MFARequiredError{}

	_ GenericAPIError = EmailNotVerifiedError{}
	_ json.Marshaler  = EmailNotVerifiedError{}
)

// MFARequiredError is returned to the client after the initial login API call if the login requires MFA.
type MFARequiredError struct{}

func (e MFARequiredError) Cause() error {
	return ErrMFARequired
}

func (e MFARequiredError) Error() string {
	return e.FriendlyMessage()
}

func (e MFARequiredError) FriendlyMessage() string {
	return e.Cause().Error()
}

func (e MFARequiredError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"error": e.FriendlyMessage(),
		"code":  "MFA_REQUIRED",
	})
}

// EmailNotVerifiedError is returned to the client when they attempt to authenticate using a login with an email that
// has not yet been verified.
type EmailNotVerifiedError struct{}

func (e EmailNotVerifiedError) Cause() error {
	return ErrEmailNotVerified
}

func (e EmailNotVerifiedError) Error() string {
	return e.FriendlyMessage()
}

func (e EmailNotVerifiedError) FriendlyMessage() string {
	return e.Cause().Error()
}

func (e EmailNotVerifiedError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"error": e.FriendlyMessage(),
		"code":  "EMAIL_NOT_VERIFIED",
	})
}

// EmailNotVerifiedError is returned to the client when they attempt to authenticate using a login with an email that
// has not yet been verified.
type EmailAlreadyExists struct{}

func (e EmailAlreadyExists) Cause() error {
	return ErrEmailAlreadyExists
}

func (e EmailAlreadyExists) Error() string {
	return e.FriendlyMessage()
}

func (e EmailAlreadyExists) FriendlyMessage() string {
	return e.Cause().Error()
}

func (e EmailAlreadyExists) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"error": e.FriendlyMessage(),
		"code":  "EMAIL_IN_USE",
	})
}

// PasswordResetRequiredError is returned to the client when they attempt to login to an account that must have its
// password updated for any reason. A short lived token is returned to the client that can be used to call the reset
// password endpoint with an updated password.
type PasswordResetRequiredError struct {
	ResetToken string `json:"resetToken"`
}

func (e PasswordResetRequiredError) Cause() error {
	return ErrPasswordChangeRequired
}

func (e PasswordResetRequiredError) Error() string {
	return e.FriendlyMessage()
}

func (e PasswordResetRequiredError) FriendlyMessage() string {
	return e.Cause().Error()
}

func (e PasswordResetRequiredError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"resetToken": e.ResetToken,
		"error":      e.FriendlyMessage(),
		"code":       "PASSWORD_CHANGE_REQUIRED",
	})
}

// SubscriptionNotActiveError is returned to the client whenever they attempt to make an API call to an endpoint that
// requires an active subscription.
type SubscriptionNotActiveError struct {
	// Will include a message indicating that the user's subscription is not active. Will always be returned with a 402
	// status code.
	Error string `json:"error" example:"subscription is not active"`
}

// MalformedJSONError is returned to the client when the request body is not valid JSON or cannot be properly parsed.
type MalformedJSONError struct {
	// Will include a message indicating that the request body is not valid JSON.
	Error string `json:"error" example:"malformed json"`
}

type ApiError struct {
	Error string `json:"error" example:"something went wrong on our end"`
}

type LinkNotFoundError struct {
	// This error is returned when the user attempts to retrieve a link that does not exist or belong to their account.
	Error string `json:"error" example:"failed to retrieve link: record does not exist"`
}

type SpendingNotFoundError struct {
	// This error is returned when the user attempts to retrieve a spending object that does not exist or belong to their account.
	Error string `json:"error" example:"failed to retrieve spending: record does not exist"`
}

type InvalidLinkIdError struct {
	// Contains an error telling the user that they must provide a valid link Id for this request.
	Error string `json:"error" example:"must specify a link Id to retrieve"`
}

type InvalidBankAccountIdError struct {
	// Contains an error telling the user that they must provide a valid bank account Id for this request.
	Error string `json:"error" example:"invalid bank account Id provided"`
}
