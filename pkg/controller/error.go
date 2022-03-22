package controller

import (
	"encoding/json"
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/pkg/errors"
)

type GenericAPIError interface {
	Cause() error
	Error() string
	FriendlyMessage() string
}

// onAnyErrorCode is the primary wrapper for handling errors and returning them in a "pretty" way to the client. It
// will also report errors to Sentry as long as that error is not for a forbidden status code. It is assumed that if
// the response has a forbidden status code that it is due to user error and not due to a problem in the API. This may
// prove to be a stupid idea in the future :tada:.
func (c *Controller) onAnyErrorCode(ctx iris.Context) {
	err := ctx.GetErr()
	if err == nil {
		return
	}

	switch ctx.GetStatusCode() {
	case http.StatusForbidden:
		// Don't report errors for forbidden status code.
	default:
		// TODO Add something to exclude some custom errors like MFA required from being reported.
		c.reportError(ctx, err)
	}

	switch actualError := err.(type) {
	case json.Marshaler:
		// If the error we have been provided implements a custom JSON formatter then pass it directly to the output, the
		// error sanitization will be handled by the error object itself and we do not need to do anything.
		ctx.JSON(actualError)
	case GenericAPIError:
		// If the error provided implements our generic error interface then we can just return the friendly messsage.
		ctx.JSON(map[string]interface{}{
			"error": actualError.FriendlyMessage(),
		})
	default:
		// This will provide backwards compatability for the time being, but in the future I would like to deprecate this
		// path entirely. If a raw error object happened to be passed this far back up the chain then we would just want to
		// print something like "internal error" or something to make sure we don't expose anything sensitive.
		// TODO (elliotcourant) Make sure errors in the future no longer fall through to this path.
		ctx.JSON(map[string]interface{}{
			"error": err.Error(),
		})
	}
}

var (
	ErrMFARequired      = errors.New("login requires MFA")
	ErrEmailNotVerified = errors.New("email address is not verified")
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
