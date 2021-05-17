package controller

import (
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/config"
	"github.com/monetrapp/rest-api/pkg/internal/myownsanity"
	"github.com/monetrapp/rest-api/pkg/models"
	"github.com/monetrapp/rest-api/pkg/swag"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"time"
)

func (c *Controller) handleBilling(p iris.Party) {
	p.Get("/plans", c.getBillingPlans)
	p.Post("/subscribe", c.postSubscribe)
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

func (c *Controller) postSubscribe(ctx iris.Context) {
	log := c.getLog(ctx)

	var subscribeRequest struct {
		PriceId         string `json:"priceId"`
		PaymentMethodId string `json:"paymentMethodId"`
	}
	if err := ctx.ReadJSON(&subscribeRequest); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	repo := c.mustGetAuthenticatedRepository(ctx)

	{ // Handle the user potentially already having a subscription.
		activeSubscription, err := repo.GetActiveSubscription(c.getContext(ctx))
		if err != nil {
			c.wrapPgError(ctx, err, "failed to validate a subscription does not already exist")
			return
		}

		if activeSubscription != nil {
			c.badRequest(ctx, "an active subscription already exists, cannot create another one")
			return
		}
	}

	var plan config.Plan
	{ // Validate the price against our configuration.
		var foundValidPlan bool
		for _, planItem := range c.configuration.Stripe.Plans {
			if planItem.StripePriceId == subscribeRequest.PriceId {
				foundValidPlan = true
				plan = planItem
			}
		}

		if !foundValidPlan {
			c.badRequest(ctx, "invalid price Id provided")
			return
		}
	}

	me, err := repo.GetMe()
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve user details")
		return
	}

	if me.StripeCustomerId == nil {
		log.Info("user does not have a stripe customer record, creating one")
		newCustomer, err := c.stripe.CreateCustomer(c.getContext(ctx), stripe.CustomerParams{
			Email: &me.Login.Email,
			Name:  stripe.String(me.Login.FirstName + " " + me.Login.LastName),
		})
		if err != nil {
			log.WithError(err).Error("could not create a stripe customer for user")
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not create a stripe customer for user")
			return
		}

		me.StripeCustomerId = &newCustomer.ID
		log.Debugf("updating user with new stripe customer Id")
		if err = repo.UpdateUser(c.getContext(ctx), me); err != nil {
			c.wrapPgError(ctx, err, "could not update user with new stripe customer Id")
			return
		}
	}

	log.Debugf("attaching payment method to stripe customer")
	paymentMethod, err := c.stripe.AttachPaymentMethod(
		c.getContext(ctx),
		subscribeRequest.PaymentMethodId,
		*me.StripeCustomerId,
	)
	if err != nil {
		log.WithError(err).Error("failed to attach payment method to stripe customer")
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not associate payment method")
		return
	}

	log.Debugf("updating customer with new default payment method")
	_, err = c.stripe.UpdateCustomer(c.getContext(ctx), *me.StripeCustomerId, stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: &paymentMethod.ID,
		},
	})
	if err != nil {
		log.WithError(err).Error("failed to update customer's default payment method")
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not set customer's default payment method")
		return
	}

	subscriptionParams := &stripe.SubscriptionParams{
		Customer:             me.StripeCustomerId,
		DefaultPaymentMethod: &paymentMethod.ID,
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan:     stripe.String(subscribeRequest.PriceId),
				Quantity: stripe.Int64(1),
			},
		},
		TrialPeriodDays: stripe.Int64(int64(plan.FreeTrialDays)),
	}
	subscriptionParams.AddExpand("latest_invoice.payment_intent")

	log.Debugf("creating subscription")
	stripeSubscription, err := c.stripe.CreateSubscription(c.getContext(ctx), *subscriptionParams)
	if err != nil {
		log.WithError(err).Error("failed to create subscription for customer")
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create subscription")
		return
	}

	var trialStart, trialEnd *time.Time
	if stripeSubscription.TrialEnd != 0 {
		trialStart = myownsanity.TimeP(time.Now().UTC())
		trialEnd = myownsanity.TimeP(time.Unix(stripeSubscription.TrialEnd, 0).UTC())
	}

	subscription := &models.Subscription{
		StripeSubscriptionId: stripeSubscription.ID,
		StripeCustomerId:     *me.StripeCustomerId,
		StripePriceId:        subscribeRequest.PriceId,
		Features:             plan.Features,
		Status:               stripeSubscription.Status,
		TrialStart:           trialStart,
		TrialEnd:             trialEnd,
	}

	if err = repo.CreateSubscription(c.getContext(ctx), subscription); err != nil {
		log.WithError(err).Error("failed to store subscription details")
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to store subscription details")
		return
	}



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
