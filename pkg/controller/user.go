package controller

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
)

func (c *Controller) handleUsers(p router.Party) {
	p.Get("/me", c.getMe)
	p.Put("/security/password")
	p.Post("/security/totp")
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
// @Failure 403 {object} ApiError There is no authenticated user.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) getMe(ctx iris.Context) {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusForbidden, "cannot retrieve user details")
		return
	}

	user, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
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

func (c *Controller) changePassword(ctx iris.Context) {
	repo, err := c.getAuthenticatedRepository(ctx)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusForbidden, "cannot retrieve user details")
		return
	}
}
