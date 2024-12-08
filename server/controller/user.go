package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/communication"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/security"
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

	claims := c.mustGetClaims(ctx)

	me := map[string]interface{}{
		"user":            user,
		"mfaPending":      false,
		"isSetup":         isSetup,
		"isActive":        true,
		"isTrialing":      false,
		"activeUntil":     nil,
		"trialingUntil":   nil,
		"hasSubscription": false,
		"supportIdentity": nil,
	}

	// If the "me" endpoint was called after they authenticated, but they still
	// need to provide MFA, then direct them to that page regardless of their
	// subscription status.
	if claims.Scope == security.MultiFactorScope {
		me["mfaPending"] = true
		me["nextUrl"] = "/login/multifactor"
	}

	if !c.Configuration.Stripe.IsBillingEnabled() {
		// When billing is not enabled we will always return the user state such that they are seen as active forever and
		// not trialing.
		return ctx.JSON(http.StatusOK, me)
	}

	// But when billing is enabled we need to handle what is basically three states.
	// - They have an active subscription (active until is in the future, or trial ends at is in the future)
	// - They have no subscription at all, or their trial has expired and they need to start one.
	// - They have a subscription but it has lapsed or has been cancelled.
	hasSubscrption := user.Account.HasSubscription()
	subscriptionIsActive := user.Account.IsSubscriptionActive(c.Clock.Now())
	subscriptionIsTrial := user.Account.IsTrialing(c.Clock.Now())

	// If billing is enabled then we want to populate these fields with real
	// values.
	me["isActive"] = subscriptionIsActive
	me["isTrialing"] = subscriptionIsTrial
	me["activeUntil"] = user.Account.SubscriptionActiveUntil
	me["trialingUntil"] = user.Account.TrialEndsAt
	me["hasSubscription"] = hasSubscrption

	if claims.Scope != security.MultiFactorScope && !subscriptionIsActive {
		// But if they are not currently required to provide MFA AND their
		// subscription is not active. Then redirect them to the account subscribe
		// endpoint.
		// TODO Make sure to implement this logic in the MFA endpoint as well!
		me["nextUrl"] = "/account/subscribe"
	}

	// If the customer support integration is enabled and they are fully
	// authenticated. Then include the end user's support identity in the
	// response.
	if claims.Scope == security.AuthenticatedScope && c.Configuration.Support.GetChatwootEnabled() {
		secret := []byte(c.Configuration.Support.ChatwootIdentityValidation)
		hash := hmac.New(sha256.New, secret)
		hash.Write([]byte(user.UserId))
		me["supportIdentity"] = hex.EncodeToString(hash.Sum(nil))
	}

	return ctx.JSON(http.StatusOK, me)
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
		if err := c.Email.SendEmail(
			c.getContext(ctx),
			communication.PasswordChangedParams{
				BaseURL:      c.Configuration.Server.GetBaseURL().String(),
				Email:        user.Login.Email,
				FirstName:    user.Login.FirstName,
				LastName:     user.Login.LastName,
				SupportEmail: "support@monetr.app",
			},
		); err != nil {
			return c.wrapAndReturnError(
				ctx,
				err,
				http.StatusInternalServerError,
				"Failed to send password changed notification",
			)
		}

		return ctx.NoContent(http.StatusOK)
	default:
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to change password")
	}
}

func (c *Controller) postSetupTOTP(ctx echo.Context) error {
	secureRepo := c.mustGetSecurityRepository(ctx)
	// Try to actually setup TOTP for the current login. This will return an error
	// if they already have it setup.
	uri, recoveryCodes, err := secureRepo.SetupTOTP(
		c.getContext(ctx),
		c.mustGetLoginId(ctx),
	)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to setup TOTP")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"uri":           uri,
		"recoveryCodes": recoveryCodes,
	})
}

func (c *Controller) postConfirmTOTP(ctx echo.Context) error {
	var request struct {
		TOTP string `json:"totp"`
	}
	secureRepo := c.mustGetSecurityRepository(ctx)
	if err := ctx.Bind(&request); err != nil {
		return c.invalidJson(ctx)
	}

	request.TOTP = strings.TrimSpace(request.TOTP)
	if request.TOTP == "" {
		return c.badRequest(ctx, "TOTP code is required")
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Unable to retrieve current user")
	}

	err = secureRepo.EnableTOTP(c.getContext(ctx), me.LoginId, request.TOTP)
	if err != nil {
		switch errors.Cause(err) {
		case models.ErrTOTPNotValid:
			return c.badRequestError(ctx, err, "Failed to enable TOTP, invalid code provided")
		default:
			return c.badRequestError(ctx, err, "Failed to enable TOTP")
		}
	}

	return ctx.NoContent(http.StatusOK)
}
