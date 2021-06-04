package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
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

	log := c.getLog(ctx).WithFields(logrus.Fields{
		"type": stripeEvent.Type,
		"id":   stripeEvent.ID,
	})

	log.Debug("received webhook")

	switch stripeEvent.Type {
	case "payment_intent.created":
	case "payment_intent.canceled":
	case "payment_intent.processing":
	case "payment_intent.payment_failed":
	case "payment_intent.requires_action":
	case "payment_intent.succeeded":
	case "payment_method.attached":
	case "payment_method.detached":
	case "invoice.created":
	case "invoice.paid":
	case "invoice.payment_failed":
	case "invoice.payment_succeeded":
	case "invoice.upcoming":
	case "customer.subscription.created":
	default:
		log.Warn("cannot handle stripe webhook event type")
	}
}
