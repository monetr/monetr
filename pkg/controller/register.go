package controller

import (
	"net/http"
	"strings"

	"github.com/kataras/iris/v12/context"
	"github.com/pkg/errors"
)

func (c *Controller) registerEndpoint(ctx *context.Context) {
	var registerRequest struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Captcha   string `json:"captcha"`
	}
	if err := ctx.ReadJSON(&registerRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid register JSON")
		return
	}

	// This will take the captcha from the request and validate it if the API is
	// configured to do so. If it is enabled and the captcha fails then an error
	// is returned to the client.
	if err := c.validateCaptchaMaybe(registerRequest.Captcha); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "valid ReCAPTCHA is required")
		return
	}

	registerRequest.Email = strings.TrimSpace(registerRequest.Email)
	registerRequest.Password = strings.TrimSpace(registerRequest.Password)
	registerRequest.FirstName = strings.TrimSpace(registerRequest.FirstName)

	if err := c.validateRegistration(
		registerRequest.Email,
		registerRequest.Password,
		registerRequest.FirstName,
	); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "invalid registration")
		return
	}

}

func (c *Controller) validateRegistration(email, password, firstName string) error {
	if len(email) == 0 {
		return errors.Errorf("email cannot be blank")
	}

	if len(password) < 8 {
		return errors.Errorf("password must be at least 8 characters")
	}

	if len(firstName) == 0 {
		return errors.Errorf("first name cannot be blank")
	}

	return nil
}

func (c *Controller) validateCaptchaMaybe(captcha string) error {
	if !c.configuration.ReCAPTCHA.Enabled {
		// If it is disabled then we don't need to do anything.
		return nil
	}

	if len(captcha) == 0 {
		return errors.Errorf("captcha is not valid")
	}

	return c.captcha.Verify(captcha)
}
