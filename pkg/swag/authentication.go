package swag

import "github.com/monetr/monetr/pkg/models"

type LoginRequest struct {
	// The email associated with our login. Is unique and case-insensitive.
	Email string `json:"email" example:"your.email@gmail.com"`
	// Your login password.
	Password string `json:"password" example:"tHEBeSTPaSsWOrdYoUCaNCOmeUpWiTH"`
	// ReCAPTCHA value from validation. Required if `verifyLogin` is enabled on the server.
	Captcha *string `json:"captcha" example:"03AGdBq266UHyZ62gfKGJozRNQz17oIhSlj9S9S..." extensions:"x-nullable"`
	// TOTP is used to provide an MFA code for the login process. It is not required to provide this code. If a login has
	// TOTP enabled then the request will fail and should be resubmitted with the TOTP code provided.
	TOTP string `json:"totp" example:"123456" extensions:"x-nullable"`
}

type LoginResponse struct {
	// A JWT that can be used to make authenticated requests for the user.
	Token string `json:"token" example:"eyJhbGciOiJI..."`
	// Indicates whether or not the user that has been authenticated has an active subscription. The UI will use this to
	// redirect the user to a payment page if their subscription is not active. If this field is not present then
	// billing is either not enabled. Or the user's subscription is active and no action needs to be taken.
	IsActive bool `json:"isActive" example:"true" extensions:"x-nullable"`
	// Next URL is provided by the API if the user needs to be redirected immediately after authenticating. This is used
	// in conjunction with the `isActive` field for directing users to the payment page gracefully. If this field is not
	// present then billing is either not enabled. Or the user's subscription is active and no action needs to be taken.
	// It is possible that this field may be used in the future independent of `isActive` so logic should be build for
	// it regardless of the `isActive` field's presence.
	NextUrl string `json:"nextUrl" example:"/account/subscribe" extensions:"x-nullable"`
}

type LoginInvalidRequestResponse struct {
	// The API performs a handful of validations on the request from the client. If any of these validations fail before
	// the credentials are even processed then an invalid request response will be returned to the client. These things
	// can be caused by not providing a valid email address (or not providing one at all), not providing a password that
	// is long enough; or not providing a valid ReCAPTCHA when it is required by the config.
	Error string `json:"error" example:"login is not valid: email address provided is not valid"`
}

type LoginInvalidCredentialsResponse struct {
	// If the client provides an invalid email address and password pair. Then the API will reject the request with an
	// error like the one here.
	Error string `json:"error" example:"invalid email and password"`
}

// LoginPreconditionRequiredResponse is returned to the client during authentication even if the credentials provided
// are valid. This is returned because something more is required in order to authenticate this login.
type LoginPreconditionRequiredResponse struct {
	// Error will contain a generic message about the problem.
	Error string `json:"error" example:"email address is not verified"`
	// Code indicates the type of precondition that is required by this endpoint.
	// * `MFA_REQUIRED` - MFA is required for this login to authenticate. The request can be remade with the MFA included
	//   in the subsequent request.
	// * `EMAIL_NOT_VERIFIED` - The email address for the provided login is not verified yet. The user must click the
	//   verify link in the email they received in order to verify that they own the email address.
	Code string `json:"code" enums:"MFA_REQUIRED,EMAIL_NOT_VERIFIED" example:"EMAIL_NOT_VERIFIED"`
}

type RegisterRequest struct {
	// The email address you want to have associated with your login and user. This is only used for verification
	// purposes like resetting a forgotten password. Or for billing. You are **never** added to any mailing list here.
	Email string `json:"email" example:"your.email@yahoo.com"`
	// Your desired login password.
	Password string `json:"password" example:"tHEBeSTPaSsWOrdYoUCaNCOmeUpWiTH"`
	// Your first name. Currently required for registration but might be able to make it optional in the future for
	// manual only registrations (not plaid linked). And people who are on a free trial.
	FirstName string `json:"firstName" example:"Doug"`
	// Your last name or "family" name. Whether or not this is required depends on the plaid configuration, when we are
	// linking bank accounts to users we do need the user's full legal name.
	LastName *string `json:"lastName" example:"Dimmadome" extensions:"x-nullable"`
	// Your timezone in the "TZ Database Name" format. This is used for determining when midnight is for funding
	// schedules to be processed for your account.
	Timezone string `json:"timezone" example:"America/Chicago"`
	// ReCAPTCHA value from validation. Required if `verifyRegistration` is enabled on the server.
	Captcha *string `json:"captcha" example:"03AGdBq266UHyZ62gfKGJozRNQz17oIhSlj9S9S..." extensions:"x-nullable"`
	// A beta code given to you to test or demo the application. This is primarily used in an environment where it would
	// cost money to link a bank account with a user. But testing against real bank accounts is necessary. So to prevent
	// anyone just creating accounts and linking their bank account for free, we use beta codes to verify that they are
	// someone who is supposed to be there. Leave this null or don't include at all if it is not required by the API
	// configuration.
	BetaCode *string `json:"betaCode" example:"F2917D98-024633A8" extensions:"x-nullable"`
}

type RegisterResponse struct {
	// This is a work in progress field, the end goal being that the API could easily direct the UI to different steps
	// based on the state of a user. If they require MFA then direct them to an MFA screen. If their subscription is
	// expired direct them to a subscription screen. But at the moment it is not used.
	NextURL string `json:"nextUrl" example:"/setup"`
	// A JWT that can be used to make authenticated requests for the newly created user.
	Token string `json:"token" example:"eyJhbGciOiJI..."`
	// The created user and some basic information. This allows the UI to skip an API call to the /users/me endpoint.
	User models.User `json:"user"`
	// Message is included if the register endpoint needs to display something to the end user. Currently this is used
	// to return a message to the user indicating that we have sent them an email to verify they own the provided email
	// address. This field is not always present.
	Message string `json:"message" example:"A verification email has been sent to your email address, please verify your email." extensions:"x-nullable"`
	// `requireVerification` is used by the UI to determine how to handle an "after-sign-up" situation. If verification
	// is not required then the UI can try to follow the path where this endpoint returns a token. If verification is
	// required then the UI should show a message telling the user to check their email.
	RequireVerification bool `json:"requireVerification" example:"true"`
}

type VerifyRequest struct {
	// A token string extracted from the URL param `?token=` from the email verification link that is sent to user's to
	// make sure that they own or at least have access to their email.
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type VerifyResponse struct {
	// Is used by the UI to direct the user at the end of the verification request. If the verification request is
	// successful (meaning the user's email address is now verified) then this will usually be `/login` to prompt the
	// user to now login using their credentials.
	NextURL string `json:"nextUrl" example:"/login"`
	// Message is used to display a toast to the user upon responding to this API call. Right now this represents a
	// successful message and tells the user they are good to go. But in the future this message could articulate
	// anything to the user.
	Message string `json:"message" example:"Your email is now verified. Please login."`
}

type ResendVerificationRequest struct {
	// Specify the email address that you want to resend the verification link to. This must be a valid email address
	// for a login that has not already validated their email.
	Email string `json:"email"`
	// ReCAPTCHA value from validation. Required if `verifyResend` is enabled on the server. This is used to prevent
	// this endpoint from easily being spammed.
	Captcha *string `json:"captcha" example:"03AGdBq266UHyZ62gfKGJozRNQz17oIhSlj9S9S..." extensions:"x-nullable"`
}

type ForgotPasswordRequest struct {
	// The email address of the login that the client want's to reset the password for. This must be a valid email
	// address.
	Email string `json:"email" example:"i.am.a.user@example.com" validate:"required"`
	// The ReCAPTCHA verification code if required by the API. This is only required if the /config endpoint indicates
	// that it is for forgot passwords.
	ReCAPTCHA *string `json:"captcha" example:"03AGdBq266UHyZ62gfKGJozRNQz17oIhSlj9S9S..." extensions:"x-nullable"`
}

type ForgotPasswordBadRequest struct {
	// Error string will be on of a few messages depending on the problem. Is used to indicate the input provided by the
	// client is not sufficient for sending a Forgot Password email.
	Error string `json:"error" example:"Must provide an email address."`
}

type ForgotPasswordEmailNotVerifiedError struct {
	// This error is returned to the client if they attempt to request a password reset link before the email has been
	// verified.
	Error string `json:"error" example:"You must verify your email before you can send forgot password requests."`
}

type ResetPasswordRequest struct {
	// The new password the client wants to use for logging in.
	Password string `json:"password" example:"superSecureP@ssword"`
	// The token that is provided to the client in an email sent by the forgot password endpoint. This is derived from
	// the `token` URL parameter in the UI for the reset password page.
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

type ResetPasswordBadRequest struct {
	Error string `json:"error" example:"Token must be provided to reset password."`
}
