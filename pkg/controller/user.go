package controller

import (
	"net/http"

	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
)

func (c *Controller) handleUsers(p router.Party) {
	p.Get("/me", c.getMe)
}

func (c *Controller) getMe(ctx *context.Context) {
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
			"user":     user,
			"isSetup":  isSetup,
			"isActive": true,
		})
		return
	}

	subscriptionIsActive, err := c.paywall.GetSubscriptionIsActive(c.getContext(ctx), user.AccountId)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not verify subscription is active")
		return
	}

	if !subscriptionIsActive {
		ctx.JSON(map[string]interface{}{
			"user":     user,
			"isSetup":  isSetup,
			"isActive": subscriptionIsActive,
			"nextUrl":  "/account/subscribe",
		})
		return
	}

	ctx.JSON(map[string]interface{}{
		"user":     user,
		"isSetup":  isSetup,
		"isActive": subscriptionIsActive,
	})
}
