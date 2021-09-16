package controller

import (
	"github.com/kataras/iris/v12/context"
)

// Application Configuration
// @Summary Get Config
// @tags Config
// @id app-config
// @description Provides the configuration that should be used by the frontend application or UI.
// @Produce json
// @Router /config [get]
// @Success 200 {object} swag.ConfigResponse
func (c *Controller) configEndpoint(ctx *context.Context) {
	type InitialPlan struct {
		Price         int64 `json:"price"`
		FreeTrialDays int32 `json:"freeTrialDays"`
	}
	var configuration struct {
		RequireLegalName    bool         `json:"requireLegalName"`
		RequirePhoneNumber  bool         `json:"requirePhoneNumber"`
		VerifyLogin         bool         `json:"verifyLogin"`
		VerifyRegister      bool         `json:"verifyRegister"`
		ReCAPTCHAKey        string       `json:"ReCAPTCHAKey,omitempty"`
		StripePublicKey     string       `json:"stripePublicKey,omitempty"`
		AllowSignUp         bool         `json:"allowSignUp"`
		AllowForgotPassword bool         `json:"allowForgotPassword"`
		LongPollPlaidSetup  bool         `json:"longPollPlaidSetup"`
		RequireBetaCode     bool         `json:"requireBetaCode"`
		InitialPlan         *InitialPlan `json:"initialPlan"`
		BillingEnabled      bool         `json:"billingEnabled"`
	}

	// If ReCAPTCHA is enabled then we want to provide the UI our public key as
	// well as whether or not we want it to verify logins and registrations.
	if c.configuration.ReCAPTCHA.Enabled {
		configuration.ReCAPTCHAKey = c.configuration.ReCAPTCHA.PublicKey
		configuration.VerifyLogin = c.configuration.ReCAPTCHA.VerifyLogin
		configuration.VerifyRegister = c.configuration.ReCAPTCHA.VerifyRegister
	}

	// We can only allow forgot password if SMTP is enabled. Otherwise we have
	// no way of sending an email to the user.
	if c.configuration.Email.Enabled {
		configuration.AllowForgotPassword = true
	}

	configuration.AllowSignUp = c.configuration.AllowSignUp

	if c.configuration.Plaid.EnableReturningUserExperience {
		configuration.RequireLegalName = true
		configuration.RequirePhoneNumber = true
	}

	if c.configuration.Stripe.Enabled {
		configuration.StripePublicKey = c.configuration.Stripe.PublicKey

		if c.configuration.Stripe.IsBillingEnabled() && c.configuration.Stripe.InitialPlan != nil {
			price, err := c.stripe.GetPriceById(
				c.getContext(ctx),
				c.configuration.Stripe.InitialPlan.StripePriceId,
			)
			if err != nil {
				c.getLog(ctx).Warn("failed to retrieve stripe price for initial plan")
			} else {
				configuration.InitialPlan = &InitialPlan{
					Price:         price.UnitAmount,
					FreeTrialDays: c.configuration.Stripe.InitialPlan.FreeTrialDays,
				}
			}
		}

		configuration.BillingEnabled = c.configuration.Stripe.BillingEnabled
	}

	configuration.RequireBetaCode = c.configuration.Beta.EnableBetaCodes

	// Just make this true for now, this might change in the future as I do websockets.
	configuration.LongPollPlaidSetup = true

	ctx.JSON(configuration)
}
