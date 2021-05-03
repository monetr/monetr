package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/stripe/stripe-go/v72/webhook"
	"net/http"
)

func (c *Controller) handleStripe(p iris.Party) {
	p.Post("/webhook", c.handleStripeWebhook)
}

func (c *Controller) handleStripeWebhook(ctx iris.Context) {
	stripeSignature := ctx.GetHeader("Stripe-Signature")
	if stripeSignature == "" {
		c.badRequest(ctx, "stripe signature is missing")
		return
	}

	requestBody, err := ctx.GetBody()
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to read request body")
		return
	}

	stripeEvent, err := webhook.ConstructEvent(requestBody, stripeSignature, c.configuration.Stripe.WebhookSecret)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to validate stripe event")
		return
	}

	c.log.Debugf("received webhook: %s", stripeEvent.ID)
}
