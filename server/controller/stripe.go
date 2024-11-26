package controller

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/stripe/stripe-go/v81/webhook"
)

func (c *Controller) handleStripeWebhook(ctx echo.Context) error {
	if !c.Configuration.Stripe.Enabled || !c.Configuration.Stripe.WebhooksEnabled {
		return c.notFound(ctx, "stripe webhooks not enabled on this server")
	}

	stripeSignature := ctx.Request().Header.Get("Stripe-Signature")
	if stripeSignature == "" {
		return c.badRequest(ctx, "stripe signature is missing")
	}

	body := ctx.Request().Body
	if body == nil {
		crumbs.IndicateBug(c.getContext(ctx), "body on request is nil for stripe webhook", nil)
		return c.badRequest(ctx, "cannot read body")
	}
	defer body.Close()

	requestBody, err := io.ReadAll(body)
	if err != nil {
		c.reportError(ctx, err)
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to read request body")
	}

	stripeEvent, err := webhook.ConstructEventWithOptions(
		requestBody,
		stripeSignature,
		c.Configuration.Stripe.WebhookSecret,
		webhook.ConstructEventOptions{
			Tolerance:                webhook.DefaultTolerance,
			IgnoreAPIVersionMismatch: true,
		},
	)
	if err != nil {
		c.reportError(ctx, err)
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to validate stripe event")
	}

	if err = c.Billing.HandleStripeWebhook(
		c.getContext(ctx),
		stripeEvent,
	); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to handle stripe webhook")
	}

	return ctx.NoContent(http.StatusOK)
}
