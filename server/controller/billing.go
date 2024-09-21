package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/billing"
	"github.com/pkg/errors"
)

// Create Checkout Session
// @Summary Create Checkout Session
// @id create-checkout-session
// @tags Billing
// @description Create a checkout session for Stripe. This is used to manage new subscriptions to monetr and offload the
// @description complexity of managing subscriptions. **Note:** You cannot create a checkout session if you already have
// @description a subscription that is not canceled associated with the account.
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param createCheckoutSession body swag.CreateCheckoutSessionRequest true "New Checkout Session"
// @Router /billing/create_checkout [post]
// @Success 200 {object} swag.CreateCheckoutSessionResponse
// @Failure 400 {object} ApiError A bad request can be returned if the account already has an active subscription, or an incomplete subscription already created.
// @Failure 500 {object} ApiError Something went wrong on our end or when communicating with Stripe.
func (c *Controller) handlePostCreateCheckout(ctx echo.Context) error {
	if !c.Configuration.Stripe.IsBillingEnabled() {
		return c.notFound(ctx, "billing is not enabled")
	}

	var request struct {
		// The path that the user should be returned to if they exit the checkout
		// session.
		CancelPath *string `json:"cancelPath"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
	}

	me, err := c.mustGetAuthenticatedRepository(ctx).GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to retrieve user details")
	}

	checkoutSession, err := c.Billing.CreateCheckout(
		c.getContext(ctx),
		*me.Login,
		c.mustGetAccountId(ctx),
		request.CancelPath,
	)
	switch errors.Cause(err) {
	case billing.ErrSubscriptionAlreadyActive:
		return c.badRequest(ctx, "There is already an active subscription for your account")
	case billing.ErrSubscriptionAlreadyExists:
		return c.badRequest(ctx, "There is already a subscription associated with your account")
	case nil:
		// Nothing
	default:
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create checkout session")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"url":       checkoutSession.URL,
		"sessionId": checkoutSession.ID,
	})
}

// Retrieve Post-Checkout Session Details
// @Summary Get Post-Checkout Session Details
// @id get-post-checkout-session-details
// @tags Billing
// @description After completing a checkout session, retrieve the outcome of the checkout session and persist it immediately.
// @Produce json
// @Security ApiKeyAuth
// @Router /billing/checkout/{checkoutSessionId} [get]
// @Param checkoutSessionId path string true "Stripe Checkout Session ID"
// @Success 200 {object} swag.AfterCheckoutResponse
// @Failure 400 {object} ApiError Invalid request.
// @Failure 500 {object} ApiError Something went wrong on our end or when communicating with Stripe.
func (c *Controller) handleGetAfterCheckout(ctx echo.Context) error {
	if !c.Configuration.Stripe.IsBillingEnabled() {
		return c.notFound(ctx, "billing is not enabled")
	}

	checkoutSessionId := ctx.Param("checkoutSessionId")
	if checkoutSessionId == "" {
		return c.badRequest(ctx, "checkout session Id is required")
	}

	active, err := c.Billing.AfterCheckout(
		c.getContext(ctx),
		c.mustGetAccountId(ctx),
		checkoutSessionId,
	)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Could not complete after checkout session")
	}

	if active {
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"nextUrl":  "/",
			"isActive": true,
		})
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"message":  "Subscription is not active.",
		"nextUrl":  "/account/subscribe",
		"isActive": false,
	})
}

func (c *Controller) getBillingPortal(ctx echo.Context) error {
	if !c.Configuration.Stripe.IsBillingEnabled() {
		return c.notFound(ctx, "billing is not enabled")
	}

	me, err := c.mustGetAuthenticatedRepository(ctx).GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve current user details")
	}

	sessionUrl, err := c.Billing.CreateBillingPortal(
		c.getContext(ctx),
		*me.Login, // Account owner? Assumed?
		c.mustGetAccountId(ctx),
	)

	if err != nil {
		if errors.Cause(err) == billing.ErrMissingSubscription {
			return c.badRequest(ctx, "account does not have a subscription")
		}
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to create new stripe portal session")
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"url": sessionUrl,
	})
}
