package swag

import "github.com/monetr/monetr/pkg/models"

type LoginRequest struct {
	// The email associated with our login. Is unique and case-insensitive.
	Email string `json:"email" example:"your.email@gmail.com"`
	// Your login password.
	Password string `json:"password" example:"tHEBeSTPaSsWOrdYoUCaNCOmeUpWiTH"`
	// ReCAPTCHA value from validation. Required if `verifyLogin` is enabled on the server.
	Captcha *string `json:"captcha" example:"03AGdBq266UHyZ62gfKGJozRNQz17oIhSlj9S9S..." extensions:"x-nullable"`
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

type LoginEmailIsNotVerifiedResponse struct {
	// When email verification is required by monetr, it is possible for the client to provide perfectly valid
	// credentials in their request and still receive an error from the API. This particular error is to let the client
	// know that a token cannot be issued until the user's email address is properly verified. In the UI the user is
	// redirected to a screen to resend the verification email if this error is returned.
	Error string `json:"error" example:"email address is not verified"`
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
