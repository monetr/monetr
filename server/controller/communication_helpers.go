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
	baseUrl := c.configuration.Server.GetBaseURL()
	verifyUrl := c.configuration.Server.GetURL("/verify/email", map[string]string{
		"token": token,
	})
	err := c.email.SendVerification(c.getContext(ctx), communication.VerifyEmailParams{
		BaseURL:      baseUrl.String(),
		Email:        login.Email,
		FirstName:    login.FirstName,
		LastName:     login.LastName,
		SupportEmail: "support@monetr.app",
		VerifyURL:    verifyUrl,
	})
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
	baseUrl := c.configuration.Server.GetBaseURL()
	resetUrl := c.configuration.Server.GetURL("/password/reset", map[string]string{
		"token": token,
	})
	err := c.email.SendPasswordReset(c.getContext(ctx), communication.PasswordResetParams{
		BaseURL:      baseUrl.String(),
		Email:        login.Email,
		FirstName:    login.FirstName,
		LastName:     login.LastName,
		SupportEmail: "support@monetr.app",
		ResetURL:     resetUrl,
	})
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
