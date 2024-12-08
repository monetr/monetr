package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/models"
)

func (c *Controller) sendVerificationEmail(
	ctx echo.Context,
	login *models.Login,
	token string,
) error {
	baseUrl := c.Configuration.Server.GetBaseURL()
	verifyUrl := c.Configuration.Server.GetURL("/verify/email", map[string]string{
		"token": token,
	})
	err := c.Email.SendEmail(
		c.getContext(ctx),
		communication.VerifyEmailParams{
			BaseURL:      baseUrl.String(),
			Email:        login.Email,
			FirstName:    login.FirstName,
			LastName:     login.LastName,
			SupportEmail: c.Configuration.Support.GetSupportEmail(),
			VerifyURL:    verifyUrl,
		},
	)
	if err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"failed to send verification email",
		)
	}

	return nil
}

func (c *Controller) sendPasswordReset(
	ctx echo.Context,
	login *models.Login,
	token string,
) error {
	baseUrl := c.Configuration.Server.GetBaseURL()
	resetUrl := c.Configuration.Server.GetURL("/password/reset", map[string]string{
		"token": token,
	})
	err := c.Email.SendEmail(
		c.getContext(ctx),
		communication.PasswordResetParams{
			BaseURL:      baseUrl.String(),
			Email:        login.Email,
			FirstName:    login.FirstName,
			LastName:     login.LastName,
			SupportEmail: c.Configuration.Support.GetSupportEmail(),
			ResetURL:     resetUrl,
		},
	)
	if err != nil {
		return c.wrapAndReturnError(
			ctx,
			err,
			http.StatusInternalServerError,
			"Failed to send password reset email",
		)
	}

	return nil
}
