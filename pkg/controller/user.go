package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
)

// Get Current User Information
// @id get-current-user
// @tags User
// @Summary Get Current User
// @description Retrieve details about the currently authenticated user. If the user is not authenticated then an error
// @description will be returned to the client. This is used by the UI to determine if the user is actually
// @description authenticated. This is because cookies are stored as HTTP only, and are not visible by the JS code in
// @description the UI.
// @Security ApiKeyAuth
// @Produce json
// @Router /users/me [get]
// @Success 200 {object} swag.MeResponse
// @Failure 401 {object} ApiError There is no authenticated user.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getMe(ctx echo.Context) error {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusUnauthorized, "cannot retrieve user details")
	}

	user, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		// Remove the cookie if this happens, it means that somehow the user may have gotten a token for a user that
		// doesn't exist?
		c.removeCookieIfPresent(ctx)
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "cannot retrieve user details")
	}

	isSetup, err := repo.GetIsSetup(c.getContext(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not determine if account is setup")
	}

	if !c.configuration.Stripe.IsBillingEnabled() {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"user":            user,
			"isSetup":         isSetup,
			"isActive":        true,
			"hasSubscription": true,
		})
	}

	subscriptionIsActive := user.Account.IsSubscriptionActive()

	if !subscriptionIsActive {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"user":            user,
			"isSetup":         isSetup,
			"isActive":        subscriptionIsActive,
			"hasSubscription": user.Account.HasSubscription(),
			"nextUrl":         "/account/subscribe",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user":            user,
		"isSetup":         isSetup,
		"isActive":        subscriptionIsActive,
		"hasSubscription": user.Account.HasSubscription(),
	})
}

// Change Password
// @Summary Change Password
// @id change-password
// @tags User
// @description Change the currently authenticated user's password. This requires that the current password be provided
// @description by the client to make sure that some actor other than the current user is not performing the action. If
// @description the request succeeds and the password is changed, no JSON will be returned; just a 200 status code.
// @Security ApiKeyAuth
// @Produce json
// @Accept json
// @Param Token body swag.ChangePasswordRequest true "Forgot Password Request"
// @Router /users/security/password [put]
// @Success 200
// @Failure 400 {object} ApiError
// @Failure 401 {object} ApiError
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) changePassword(ctx echo.Context) error {
	var changePasswordRequest struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := ctx.Bind(&changePasswordRequest); err != nil {
		return c.invalidJson(ctx)
	}

	changePasswordRequest.CurrentPassword = strings.TrimSpace(changePasswordRequest.CurrentPassword)
	changePasswordRequest.NewPassword = strings.TrimSpace(changePasswordRequest.NewPassword)

	if len(changePasswordRequest.NewPassword) < 8 {
		return c.badRequest(ctx, "new password is not valid")
	}

	if changePasswordRequest.NewPassword == changePasswordRequest.CurrentPassword {
		return c.badRequest(ctx, "new password must be different from the current password")
	}

	user, err := c.mustGetAuthenticatedRepository(ctx).GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve current user details")
	}

	secureRepo := c.mustGetSecurityRepository(ctx)
	err = secureRepo.ChangePassword(
		c.getContext(ctx),
		user.LoginId,
		changePasswordRequest.CurrentPassword,
		changePasswordRequest.NewPassword,
	)
	switch errors.Cause(err) {
	case repository.ErrInvalidCredentials:
		return c.returnError(ctx, http.StatusUnauthorized, "current password provided is not correct")
	case nil:
		return ctx.NoContent(http.StatusOK)
	default:
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to change password")
	}
}
