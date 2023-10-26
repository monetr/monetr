package controller

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/stripe/stripe-go/v72/webhook"
)

func (c *Controller) handleStripeWebhook(ctx echo.Context) error {
	if !c.configuration.Stripe.Enabled || !c.configuration.Stripe.WebhooksEnabled {
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

	requestBody, err := ioutil.ReadAll(body)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to read request body")
	}

	stripeEvent, err := webhook.ConstructEvent(requestBody, stripeSignature, c.configuration.Stripe.WebhookSecret)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to validate stripe event")
	}

	if err = c.stripeWebhooks.HandleWebhook(c.getContext(ctx), stripeEvent); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to handle stripe webhook")
	}

	return ctx.NoContent(http.StatusOK)
}
