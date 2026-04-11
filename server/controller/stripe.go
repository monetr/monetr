package controller

import (
	"io"
	"math"
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

	// Not using [http.MaxBytesReader] since it leverages the response writer as
	// well. I don't want to interfere with echo's response handling or any other
	// middleware I have setup so I'm using [io.LimitReader] instead to achieve
	// the same thing. Max of 65kb based on:
	// https://github.com/stripe/stripe-go/blob/395614cfd3891376de57411afe8e02ab1f614cf3/webhook/client_handler_test.go#L14-L17
	requestBody, err := io.ReadAll(io.LimitReader(body, int64(math.MaxUint16)))
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
