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
}
