package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/monetr/monetr/pkg/build"
)

// Application Configuration
// @Summary Get Config
// @tags Config
// @id app-config
// @description Provides the configuration that should be used by the frontend application or UI.
// @Produce json
// @Router /config [get]
// @Success 200 {object} swag.ConfigResponse
func (c *Controller) configEndpoint(ctx iris.Context) {
	type InitialPlan struct {
		Price         int64 `json:"price"`
		FreeTrialDays int32 `json:"freeTrialDays"`
	}
	var configuration struct {
		RequireLegalName    bool         `json:"requireLegalName"`
		RequirePhoneNumber  bool         `json:"requirePhoneNumber"`
		VerifyLogin         bool         `json:"verifyLogin"`
		VerifyRegister      bool         `json:"verifyRegister"`
		VerifyEmailAddress  bool         `json:"verifyEmailAddress"`
		ReCAPTCHAKey        string       `json:"ReCAPTCHAKey,omitempty"`
		StripePublicKey     string       `json:"stripePublicKey,omitempty"`
		AllowSignUp         bool         `json:"allowSignUp"`
		AllowForgotPassword bool         `json:"allowForgotPassword"`
		LongPollPlaidSetup  bool         `json:"longPollPlaidSetup"`
		RequireBetaCode     bool         `json:"requireBetaCode"`
		InitialPlan         *InitialPlan `json:"initialPlan"`
		BillingEnabled      bool         `json:"billingEnabled"`
		Release             string       `json:"release"`
		Revision            string       `json:"revision"`
	}

	configuration.Release = build.Release
	configuration.Revision = build.Revision

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

	configuration.VerifyEmailAddress = c.configuration.Email.ShouldVerifyEmails()

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

// Get Public Sentry DSN
// @Summary Get Public Sentry DSN
// @tags Config
// @id get-ui-sentry-dsn
// @description Is used to allow the Sentry DSN for the UI to be configurable at runtime. This endpoint is only
// @description accessible when Sentry is enabled. The DSN it returns is the public DSN only. More information about how
// @description the DSN works can be found here:
// @description https://docs.sentry.io/product/sentry-basics/dsn-explainer/#dsn-utilization
// @Produce json
// @Router /sentry [get]
// @Success 200 {object} swag.SentryDSNResponse
func (c *Controller) getSentryUI(ctx iris.Context) {
	ctx.JSON(map[string]interface{}{
		"dsn": c.configuration.Sentry.DSN,
	})
}
