package controller

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/kataras/iris/v12"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/webhook"
	"net/http"
	"strings"
)

func (c *Controller) handleStripe(p iris.Party) {
	c.provisionStripeWebhooks()

	p.Post("/webhook", c.handleStripeWebhook)
}

func (c *Controller) provisionStripeWebhooks() {
	log := c.log.WithField("module", "stripe")
	log.Debug("provisioning stripe webhooks, checking for existing webhooks")

	var foundEndpoint *stripe.WebhookEndpoint

	endpointsIterator := c.stripeClient.WebhookEndpoints.List(&stripe.WebhookEndpointListParams{})
WebhookLoop:
	for {
		if err := endpointsIterator.Err(); err != nil {
			log.WithError(err).Error("failed to iterate over stripe webhook endpoints")
			sentry.CaptureException(err)
			return
		}

		for _, endpoint := range endpointsIterator.WebhookEndpointList().Data {
			if endpoint.Status != "enabled" || endpoint.Deleted {
				continue
			}

			if endpoint.Metadata == nil {
				continue
			}

			if environment, ok := endpoint.Metadata["environment"]; ok {
				if strings.ToLower(environment) == strings.ToLower(c.configuration.Environment) {
					foundEndpoint = endpoint
					break WebhookLoop
				}
			}
		}

		if !endpointsIterator.Next() {
			break WebhookLoop
		}
	}

	if foundEndpoint == nil {
		log.Info("no webhook endpoint was found for this instance, a new one will be created")

		description := fmt.Sprintf("REST API [%s]", strings.ToUpper(c.configuration.Environment))
		thing := "*"
		webhookUrl := fmt.Sprintf("%s/stripe/webhook", c.configuration.Stripe.WebhooksDomain)

		result, err := c.stripeClient.WebhookEndpoints.New(&stripe.WebhookEndpointParams{
			Params: stripe.Params{
				Metadata: map[string]string{
					"environment": strings.ToLower(c.configuration.Environment),
				},
			},
			Description: &description,
			EnabledEvents: []*string{
				&thing,
			},
			URL: &webhookUrl,
		})
		if err != nil {
			log.WithError(err).Error("failed to setup new webhook endpoint for rest api")
			sentry.CaptureException(err)
			return
		}

		log.Infof("successfully registered new webhook endpoint (%s) at: %s", result.ID, webhookUrl)
	} else {
		log.Debug("found existing webhook endpoint, will check for updates")
	}
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

	stripeEvent, err := webhook.ConstructEvent(requestBody, stripeSignature, "")
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "failed to validate stripe event")
		return
	}

	c.log.Debugf("received webhook: %s", stripeEvent.ID)
}
