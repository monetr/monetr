package swag

type ConfigResponse struct {
	// Indicates whether or not the UI should collect just a simple "firstName" or should require that the user provide
	// both their first and last name during registration.
	RequireLegalName bool `json:"requireLegalName"`

	// **WIP** Not currently used. This is meant to be used for doing additional verification of the user's identity to
	// streamline the bank account linking process.
	RequirePhoneNumber bool `json:"requirePhoneNumber"`

	// Tells the API client that a ReCAPTCHA verification key will be required for login API calls.
	VerifyLogin bool `json:"verifyLogin"`

	// Tells the API client that a ReCAPTCHA verification key will be required for registering a new user.
	VerifyRegister bool `json:"verifyRegister"`

	// The public ReCAPTCHA key that should be used by the frontend to verify some requests. Is omitted if ReCAPTCHA is
	// not enabled.
	ReCAPTCHAKey string `json:"ReCAPTCHAKey"`

	// The public key for Stripe, will be used for stripe elements on the frontend. Is omitted if stripe is not enabled.
	StripePublicKey string `json:"stripePublicKey"`

	// Tells the UI whether or not registration requests will be accepted by the UI.
	AllowSignUp bool `json:"allowSignUp"`

	// **WIP** Not currently used. Will be implemented once proper email verification is working. Will also require that
	// the API can send emails to the end user.
	AllowForgotPassword bool `json:"allowForgotPassword"`

	// Indicates that registration requests will require a one time use beta code in order to be accepted. Beta codes
	// must be generated before hand by an admin.
	RequireBetaCode bool `json:"requireBetaCode"`
}
