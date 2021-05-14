package controller

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/internal/myownsanity"
	"github.com/monetrapp/rest-api/pkg/swag"
	"github.com/stripe/stripe-go/v72"
	"net/http"
)

func (c *Controller) handleBilling(p iris.Party) {
	p.Get("/plans", c.getBillingPlans)
	p.Post("/create_checkout", c.handlePostCreateCheckout)
}

type Plan struct {
	Id            string                        `json:"id"`
	Name          string                        `json:"name"`
	Description   string                        `json:"description"`
	UnitPrice     int64                         `json:"unitPrice"`
	Interval      stripe.PriceRecurringInterval `json:"interval"`
	IntervalCount int64                         `json:"intervalCount"`
	FreeTrialDays int32                         `json:"freeTrialDays"`
	Active        bool                          `json:"active"`
}

func (c *Controller) getBillingPlans(ctx iris.Context) {
	log := c.getLog(ctx)

	configuredPlans := c.configuration.Stripe.Plans
	stripePriceIds := make([]string, len(configuredPlans))
	for i, plan := range configuredPlans {
		stripePriceIds[i] = plan.StripePriceId
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	subscription, err := repo.GetActiveSubscription(c.getContext(ctx))
	if err != nil {
		c.wrapPgError(ctx, err, "failed to check for active subscription")
		return
	}

	// If there is a subscription active for the current user and their price is not in our configuration file then that
	// means they might be on an old price. We will want to add this to our list to retrieve details for.
	if subscription != nil && !myownsanity.SliceContains(stripePriceIds, subscription.StripePriceId) {
		log.Debug("account has an old price, will retrieve an additional price from stripe")
		stripePriceIds = append(stripePriceIds, subscription.StripePriceId)
	}

	log.Debugf("retrieving %d price(s) from stripe", len(stripePriceIds))
	stripePrices, err := c.stripe.GetPricesById(c.getContext(ctx), stripePriceIds)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve price details from stripe")
		return
	}

	log.Debugf("retrieved %d price(s) from stripe", len(stripePrices))
	stripeProductIds := make([]string, 0, len(stripePrices))
	linq.From(stripePrices).
		SelectT(func(price stripe.Price) string {
			return price.Product.ID
		}).
		Distinct().
		ToSlice(&stripeProductIds)

	log.Debugf("retrieving %d product(s) from stripe", len(stripeProductIds))
	stripeProducts, err := c.stripe.GetProductsById(c.getContext(ctx), stripeProductIds)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve product details from stripe")
		return
	}

	plans := make([]Plan, len(stripePriceIds))
	linq.From(stripePrices).
		JoinT(
			linq.From(stripeProducts),
			func(price stripe.Price) string {
				return price.Product.ID
			},
			func(product stripe.Product) string {
				return product.ID
			},
			func(price stripe.Price, product stripe.Product) Plan {
				configPlan := linq.From(configuredPlans).
					FirstWithT(func(plan config.Plan) bool {
						return plan.StripePriceId == price.ID
					}).(config.Plan)

				return Plan{
					Id:            price.ID,
					Name:          product.Name,
					Description:   product.Description,
					UnitPrice:     price.UnitAmount,
					Interval:      price.Recurring.Interval,
					IntervalCount: price.Recurring.IntervalCount,
					FreeTrialDays: configPlan.FreeTrialDays,
					Active:        subscription != nil && subscription.StripePriceId == price.ID,
				}
			},
		).
		ToSlice(&plans)

	ctx.JSON(plans)
	return
}

// Create Checkout Session
// @Summary Create Checkout Session
// @id create-checkout-session
// @tags Billing
// @description Create a checkout session for the user to enter billing information in and for us to associate it with a new subscription object.
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param createCheckoutSession body swag.CreateCheckoutSessionRequest true "New Checkout Session"
// @Router /billing/create_checkout [post]
// @Success 200 {object} swag.CreateCheckoutSessionResponse
// @Failure 400 {object} ApiError Invalid request.
// @Failure 500 {object} ApiError Something went wrong on our end or when communicating with Stripe.
func (c *Controller) handlePostCreateCheckout(ctx iris.Context) {
	var request swag.CreateCheckoutSessionRequest
	if err := ctx.ReadJSON(&request); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)
	me, err := repo.GetMe()
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve current user details")
		return
	}

	cancelUrl := fmt.Sprintf("https://%s/billing/cancel", c.configuration.UIDomainName)
	successUrl := fmt.Sprintf("https://%s/billing", c.configuration.UIDomainName)
	var email *string
	if me.Login != nil {
		email = &me.Login.Email
	}

	// TODO Lookup the stripe price using the provided price Id.
	var priceId string

	checkoutParams := &stripe.CheckoutSessionParams{
		SuccessURL:        &successUrl,
		CancelURL:         &cancelUrl,
		ClientReferenceID: nil,
		Customer:          me.StripeCustomerId,
		CustomerEmail:     email,
		Discounts:         nil,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				// Number of bank accounts?
				Amount:   nil,
				Quantity: nil,
				Price:    &priceId,
			},
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode:                      stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		PaymentIntentData:         nil,
		PaymentMethodOptions:      nil,
		SetupIntentData:           nil,
		ShippingAddressCollection: nil,
		ShippingRates:             nil,
		SubmitType:                nil,
		SubscriptionData:          nil,
	}

	result, err := c.stripeClient.CheckoutSessions.New(checkoutParams)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create checkout session")
		return
	}

	ctx.JSON(swag.CreateCheckoutSessionResponse{
		SessionId: result.ID,
	})
}
