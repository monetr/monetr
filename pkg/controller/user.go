package controller

import (
	"net/http"
	"strings"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
	"github.com/monetr/monetr/pkg/hash"
	"github.com/monetr/monetr/pkg/repository"
	"github.com/pkg/errors"
)

func (c *Controller) handleUsers(p router.Party) {
	p.Get("/me", c.getMe)
	p.Put("/security/password", c.changePassword)
}

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
func (c *Controller) getMe(ctx iris.Context) {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusUnauthorized, "cannot retrieve user details")
		return
	}

	user, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		// Remove the cookie if this happens, it means that somehow the user may have gotten a token for a user that
		// doesn't exist?
		c.removeCookieIfPresent(ctx)
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "cannot retrieve user details")
		return
	}

	isSetup, err := repo.GetIsSetup(c.getContext(ctx))
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not determine if account is setup")
		return
	}

	if !c.configuration.Stripe.IsBillingEnabled() {
		ctx.JSON(map[string]interface{}{
			"user":            user,
			"isSetup":         isSetup,
			"isActive":        true,
			"hasSubscription": true,
		})
		return
	}

	subscriptionIsActive := user.Account.IsSubscriptionActive()

	if !subscriptionIsActive {
		ctx.JSON(map[string]interface{}{
			"user":            user,
			"isSetup":         isSetup,
			"isActive":        subscriptionIsActive,
			"hasSubscription": user.Account.HasSubscription(),
			"nextUrl":         "/account/subscribe",
		})
		return
	}

	ctx.JSON(map[string]interface{}{
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
func (c *Controller) changePassword(ctx iris.Context) {
	var changePasswordRequest struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := ctx.ReadJSON(&changePasswordRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed json")
		return
	}

	changePasswordRequest.CurrentPassword = strings.TrimSpace(changePasswordRequest.CurrentPassword)
	changePasswordRequest.NewPassword = strings.TrimSpace(changePasswordRequest.NewPassword)

	if len(changePasswordRequest.NewPassword) < 8 {
		c.badRequest(ctx, "new password is not valid")
		return
	}

	if changePasswordRequest.NewPassword == changePasswordRequest.CurrentPassword {
		c.badRequest(ctx, "new password must be different from the current password")
		return
	}

	user, err := c.mustGetAuthenticatedRepository(ctx).GetMe(c.getContext(ctx))
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve current user details")
		return
	}

	oldHashedPassword := hash.HashPassword(user.Login.Email, changePasswordRequest.CurrentPassword)
	newHashedPassword := hash.HashPassword(user.Login.Email, changePasswordRequest.NewPassword)

	secureRepo := c.mustGetSecurityRepository(ctx)
	err = secureRepo.ChangePassword(c.getContext(ctx), user.LoginId, oldHashedPassword, newHashedPassword)
	switch errors.Cause(err) {
	case repository.ErrInvalidCredentials:
		c.returnError(ctx, http.StatusUnauthorized, "current password provided is not correct")
	case nil:
		ctx.StatusCode(http.StatusOK)
	default:
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to change password")
	}
}
