package controller

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/core/router"
	"github.com/stripe/stripe-go/v72"
	"net/http"
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

	user, err := repo.GetMe()
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "cannot retrieve user details")
		return
	}

	isSetup, err := repo.GetIsSetup()
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not determine if account is setup")
		return
	}

	if !c.configuration.Stripe.BillingEnabled {
		ctx.JSON(map[string]interface{}{
			"user":     user,
			"isSetup":  isSetup,
			"isActive": true,
		})
		return
	}

	subscription, err := repo.GetActiveSubscription(c.getContext(ctx))
	if err != nil {
		c.wrapPgError(ctx, err, "failed to get active subscription")
		return
	}

	var subscriptionIsActive bool
	if subscription == nil {
		subscriptionIsActive = false
	} else {
		switch subscription.Status {
		case stripe.SubscriptionStatusActive,
			stripe.SubscriptionStatusTrialing:
			subscriptionIsActive = true
		case stripe.SubscriptionStatusPastDue,
			stripe.SubscriptionStatusUnpaid,
			stripe.SubscriptionStatusCanceled,
			stripe.SubscriptionStatusIncomplete,
			stripe.SubscriptionStatusIncompleteExpired:
			subscriptionIsActive = false
		default:
			sentry.CaptureMessage(fmt.Sprintf("invalid subscription status: %s", subscription.Status))
			c.returnError(ctx, http.StatusNotImplemented, "invalid subscription status, create a github issue")
			return
		}
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
