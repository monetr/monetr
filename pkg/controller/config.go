package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/icons"
)

func (c *Controller) configEndpoint(ctx echo.Context) error {
	type InitialPlan struct {
		Price         int64 `json:"price"`
		FreeTrialDays int32 `json:"freeTrialDays"`
	}
	var configuration struct {
		RequireLegalName     bool         `json:"requireLegalName"`
		RequirePhoneNumber   bool         `json:"requirePhoneNumber"`
		VerifyLogin          bool         `json:"verifyLogin"`
		VerifyRegister       bool         `json:"verifyRegister"`
		VerifyEmailAddress   bool         `json:"verifyEmailAddress"`
		VerifyForgotPassword bool         `json:"verifyForgotPassword"`
		ReCAPTCHAKey         string       `json:"ReCAPTCHAKey,omitempty"`
		AllowSignUp          bool         `json:"allowSignUp"`
		AllowForgotPassword  bool         `json:"allowForgotPassword"`
		LongPollPlaidSetup   bool         `json:"longPollPlaidSetup"`
		RequireBetaCode      bool         `json:"requireBetaCode"`
		InitialPlan          *InitialPlan `json:"initialPlan"`
		BillingEnabled       bool         `json:"billingEnabled"`
		IconsEnabled         bool         `json:"iconsEnabled"`
		Release              string       `json:"release"`
		Revision             string       `json:"revision"`
		BuildType            string       `json:"buildType"`
		BuildTime            string       `json:"buildTime"`
	}

	configuration.Release = build.Release
	configuration.Revision = build.Revision
	configuration.BuildType = build.BuildType
	configuration.BuildTime = build.BuildTime

	// If ReCAPTCHA is enabled then we want to provide the UI our public key as
	// well as whether or not we want it to verify logins and registrations.
	if c.configuration.ReCAPTCHA.Enabled {
		configuration.ReCAPTCHAKey = c.configuration.ReCAPTCHA.PublicKey
		configuration.VerifyLogin = c.configuration.ReCAPTCHA.VerifyLogin
		configuration.VerifyRegister = c.configuration.ReCAPTCHA.VerifyRegister
	}

	// We can only allow forgot password if SMTP is enabled. Otherwise we have
	// no way of sending an email to the user.
	if c.configuration.Email.AllowPasswordReset() {
		configuration.AllowForgotPassword = true
		configuration.VerifyForgotPassword = c.configuration.ReCAPTCHA.ShouldVerifyForgotPassword()
	}

	configuration.VerifyEmailAddress = c.configuration.Email.ShouldVerifyEmails()

	configuration.AllowSignUp = c.configuration.AllowSignUp

	if c.configuration.Plaid.EnableReturningUserExperience {
		configuration.RequireLegalName = true
		configuration.RequirePhoneNumber = true
	}

	if c.configuration.Stripe.Enabled {
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

	configuration.IconsEnabled = icons.GetIconsEnabled()

	return ctx.JSON(http.StatusOK, configuration)
}

func (c *Controller) getSentryUI(ctx echo.Context) error {
	if !c.configuration.Sentry.ExternalSentryEnabled() {
		return c.notFound(ctx, "public sentry key is not enabled")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"dsn": c.configuration.Sentry.ExternalDSN,
	})
}
