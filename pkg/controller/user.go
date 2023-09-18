package controller

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
)

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
			"isTrialing":      false,
			"activeUntil":     nil,
			"hasSubscription": true,
		})
	}

	subscriptionIsActive := user.Account.IsSubscriptionActive()
	subscriptionIsTrial := user.Account.IsTrialing()

	if !subscriptionIsActive {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"user":            user,
			"isSetup":         isSetup,
			"isActive":        subscriptionIsActive,
			"isTrialing":      subscriptionIsTrial,
			"activeUntil":     user.Account.SubscriptionActiveUntil,
			"hasSubscription": user.Account.HasSubscription(),
			"nextUrl":         "/account/subscribe",
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user":            user,
		"isSetup":         isSetup,
		"isActive":        subscriptionIsActive,
		"isTrialing":      subscriptionIsTrial,
		"activeUntil":     user.Account.SubscriptionActiveUntil,
		"hasSubscription": user.Account.HasSubscription(),
	})
}

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
