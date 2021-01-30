package controller

import (
	"github.com/kataras/iris/v12/context"
)

func (c *Controller) configEndpoint(ctx *context.Context) {
	var config struct {
		VerifyLogin         bool   `json:"verifyLogin"`
		VerifyRegister      bool   `json:"verifyRegister"`
		ReCAPTCHAKey        string `json:"ReCAPTCHAKey,omitempty"`
		AllowSignUp         bool   `json:"allowSignUp"`
		AllowForgotPassword bool   `json:"allowForgotPassword"`
	}

	// If ReCAPTCHA is enabled then we want to provide the UI our public key as
	// well as whether or not we want it to verify logins and registrations.
	if c.configuration.ReCAPTCHA.Enabled {
		config.ReCAPTCHAKey = c.configuration.ReCAPTCHA.PublicKey
		config.VerifyLogin = c.configuration.ReCAPTCHA.VerifyLogin
		config.VerifyRegister = c.configuration.ReCAPTCHA.VerifyRegister
	}

	// We can only allow forgot password if SMTP is enabled. Otherwise we have
	// no way of sending an email to the user.
	if c.configuration.SMTP.Enabled {
		config.AllowForgotPassword = true
	}

	config.AllowSignUp = c.configuration.AllowSignUp

	ctx.JSON(config)
}
