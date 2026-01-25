package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/datasources/lunch_flow"
	"github.com/monetr/monetr/server/icons"
)

func (c *Controller) configEndpoint(ctx echo.Context) error {
	type InitialPlan struct {
		Price int64 `json:"price"`
	}
	var configuration struct {
		RequireLegalName       bool         `json:"requireLegalName"`
		RequirePhoneNumber     bool         `json:"requirePhoneNumber"`
		VerifyLogin            bool         `json:"verifyLogin"`
		VerifyRegister         bool         `json:"verifyRegister"`
		VerifyEmailAddress     bool         `json:"verifyEmailAddress"`
		VerifyForgotPassword   bool         `json:"verifyForgotPassword"`
		ReCAPTCHAKey           string       `json:"ReCAPTCHAKey,omitempty"`
		AllowSignUp            bool         `json:"allowSignUp"`
		AllowForgotPassword    bool         `json:"allowForgotPassword"`
		LongPollPlaidSetup     bool         `json:"longPollPlaidSetup"`
		RequireBetaCode        bool         `json:"requireBetaCode"`
		InitialPlan            *InitialPlan `json:"initialPlan"`
		BillingEnabled         bool         `json:"billingEnabled"`
		IconsEnabled           bool         `json:"iconsEnabled"`
		PlaidEnabled           bool         `json:"plaidEnabled"`
		LunchFlowEnabled       bool         `json:"lunchFlowEnabled"`
		LunchFlowDefaultAPIURL string       `json:"lunchFlowDefaultAPIURL"`
		ManualEnabled          bool         `json:"manualEnabled"`
		UploadsEnabled         bool         `json:"uploadsEnabled"`
		Release                string       `json:"release"`
		Revision               string       `json:"revision"`
		BuildType              string       `json:"buildType"`
		BuildTime              string       `json:"buildTime"`
	}

	configuration.Release = build.Release
	configuration.Revision = build.Revision
	configuration.BuildType = build.BuildType
	configuration.BuildTime = build.BuildTime

	// If ReCAPTCHA is enabled then we want to provide the UI our public key as
	// well as whether or not we want it to verify logins and registrations.
	if c.Configuration.ReCAPTCHA.Enabled {
		configuration.ReCAPTCHAKey = c.Configuration.ReCAPTCHA.PublicKey
		configuration.VerifyLogin = c.Configuration.ReCAPTCHA.VerifyLogin
		configuration.VerifyRegister = c.Configuration.ReCAPTCHA.VerifyRegister
	}

	// We can only allow forgot password if SMTP is enabled. Otherwise we have
	// no way of sending an email to the user.
	if c.Configuration.Email.AllowPasswordReset() {
		configuration.AllowForgotPassword = true
		configuration.VerifyForgotPassword = c.Configuration.ReCAPTCHA.ShouldVerifyForgotPassword()
	}

	configuration.VerifyEmailAddress = c.Configuration.Email.ShouldVerifyEmails()

	configuration.AllowSignUp = c.Configuration.AllowSignUp

	if c.Configuration.Plaid.EnableReturningUserExperience {
		configuration.RequireLegalName = true
		configuration.RequirePhoneNumber = true
	}

	configuration.BillingEnabled = c.Configuration.Stripe.IsBillingEnabled()
	if c.Configuration.Stripe.IsBillingEnabled() && c.Configuration.Stripe.InitialPlan != nil {
		price, err := c.Stripe.GetPriceById(
			c.getContext(ctx),
			c.Configuration.Stripe.InitialPlan.StripePriceId,
		)
		if err != nil {
			c.getLog(ctx).Warn("failed to retrieve stripe price for initial plan")
		} else {
			configuration.InitialPlan = &InitialPlan{
				Price: price.UnitAmount,
			}
		}
	}

	configuration.RequireBetaCode = c.Configuration.Beta.EnableBetaCodes

	// Just make this true for now, this might change in the future as I do websockets.
	configuration.LongPollPlaidSetup = true

	configuration.IconsEnabled = icons.GetIconsEnabled()
	configuration.PlaidEnabled = c.Configuration.Plaid.GetEnabled()
	configuration.ManualEnabled = true
	configuration.UploadsEnabled = c.Configuration.Storage.Enabled

	configuration.LunchFlowEnabled = c.Configuration.LunchFlow.Enabled
	configuration.LunchFlowDefaultAPIURL = lunch_flow.DefaultAPIURL()

	return ctx.JSON(http.StatusOK, configuration)
}
