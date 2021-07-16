package controller

import (
	"fmt"
	"github.com/monetr/rest-api/pkg/crumbs"
	"net/http"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/monetr/rest-api/pkg/build"
	"github.com/monetr/rest-api/pkg/config"
	"github.com/monetr/rest-api/pkg/swag"
	"github.com/stripe/stripe-go/v72"
)

func (c *Controller) handleBilling(p iris.Party) {
	p.Post("/create_checkout", c.handlePostCreateCheckout)
	p.Get("/portal", c.handleGetStripePortal)
	p.Get("/wait", c.waitForSubscription)
}

// Create Checkout Session
// @Summary Create Checkout Session
// @id create-checkout-session
// @tags Billing
// @description Create a checkout session for Stripe. This is used to manage new subscriptions to monetr and offload the complexity of managing subscriptions. **Note:** You cannot create a checkout session if you have an active subscrption.
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param createCheckoutSession body swag.CreateCheckoutSessionRequest true "New Checkout Session"
// @Router /billing/create_checkout [post]
// @Success 200 {object} swag.CreateCheckoutSessionResponse
// @Failure 400 {object} ApiError Invalid request.
// @Failure 500 {object} ApiError Something went wrong on our end or when communicating with Stripe.
func (c *Controller) handlePostCreateCheckout(ctx iris.Context) {
	isActive, err := c.paywall.GetSubscriptionIsActive(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to verify there is not already an active subscription")
		return
	}

	// If the customer already has an active subscription we do not want them to try to use this at this time. I don't
	// know how stripe handles this off the top of my head at the time of writing this. But I've setup an endpoint to
	// manage the subscriptions that already exist via the stripe portal. So existing subscriptions should be managed
	// there instead.
	if isActive {
		c.badRequest(ctx, "there is already an active subscription for your account")
		return
	}

	var request swag.CreateCheckoutSessionRequest
	if err := ctx.ReadJSON(&request); err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
		return
	}

	log := c.getLog(ctx)

	var plan config.Plan
	if request.PriceId == "" && c.configuration.Stripe.InitialPlan != nil {
		request.PriceId = c.configuration.Stripe.InitialPlan.StripePriceId
		plan = *c.configuration.Stripe.InitialPlan
	} else {
		if request.PriceId == "" {
			c.badRequest(ctx, "must provide a price id")
			return
		}

		{ // Validate the price against our configuration.
			var foundValidPlan bool
			for _, planItem := range c.configuration.Stripe.Plans {
				if planItem.StripePriceId == request.PriceId {
					foundValidPlan = true
					plan = planItem
					break
				}
			}

			if !foundValidPlan {
				c.badRequest(ctx, "invalid price Id provided")
				return
			}
		}
	}

	crumbs.Debug(c.getContext(ctx), "Creating checkout session for price", map[string]interface{}{
		"priceId":       plan.StripePriceId,
		"freeTrialDays": plan.FreeTrialDays,
	})

	repo := c.mustGetAuthenticatedRepository(ctx)

	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve account")
		return
	}

	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve current user details")
		return
	}

	// Check to see if the account does not already have a stripe customer Id. If they don't have one then we want to
	// create one.
	if account.StripeCustomerId == nil {
		crumbs.Debug(c.getContext(ctx), "Account does not have a Stripe Customer ID, a customer will be created.", nil)
		log.Warn("attempting to create a checkout session for an account with no customer, customer will be created")
		name := me.FirstName + " " + me.LastName
		customer, err := c.stripe.CreateCustomer(c.getContext(ctx), stripe.CustomerParams{
			Email: &me.Login.Email,
			Name:  &name,
			Params: stripe.Params{
				Metadata: map[string]string{
					"environment": c.configuration.Environment,
					"revision":    build.Revision,
					"release":     build.Release,
				},
			},
		})
		if err != nil {
			log.WithError(err).Error("failed to create stripe customer for checkout")
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create stripe customer")
			return
		}

		account.StripeCustomerId = &customer.ID
		if err = c.accounts.UpdateAccount(c.getContext(ctx), account); err != nil {
			log.WithError(err).Error("failed to update account with new customer Id")
			c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to update account with new customer Id")
			return
		}

		log.Info("successfully created stripe customer for account")
	}

	successUrl := fmt.Sprintf("https://%s/account/subscribe/after?session={CHECKOUT_SESSION_ID}", c.configuration.UIDomainName)
	cancelUrl := fmt.Sprintf("https://%s/account/subscribe", c.configuration.UIDomainName)
	if request.CancelPath != nil {
		cancelUrl = fmt.Sprintf("https://%s%s", c.configuration.UIDomainName, *request.CancelPath)
	}

	crumbs.Debug(c.getContext(ctx), "Creating Stripe Checkout Session", map[string]interface{}{
		"successUrl": successUrl,
		"cancelUrl":  cancelUrl,
	})

	checkoutParams := &stripe.CheckoutSessionParams{
		SuccessURL: &successUrl,
		CancelURL:  &cancelUrl,
		Customer:   account.StripeCustomerId,
		Discounts:  nil,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				// Number of bank accounts?
				Amount:   nil,
				Quantity: stripe.Int64(1),
				Price:    &plan.StripePriceId,
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
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Coupon:          nil,
			DefaultTaxRates: nil,
			Metadata: map[string]string{
				"environment": c.configuration.Environment,
				"revision":    build.Revision,
				"release":     build.Release,
			},
			TransferData:    nil,
			TrialEnd:        nil,
			TrialFromPlan:   nil,
			TrialPeriodDays: nil,
		},
	}

	if plan.FreeTrialDays > 0 {
		checkoutParams.SubscriptionData.TrialPeriodDays = stripe.Int64(int64(plan.FreeTrialDays))
	}

	result, err := c.stripe.NewCheckoutSession(c.getContext(ctx), checkoutParams)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create checkout session")
		return
	}

	ctx.JSON(swag.CreateCheckoutSessionResponse{
		SessionId: result.ID,
	})
}

// Get Stripe Portal
// @id get-stripe-portal
// @tags Billing
// @Summary Get Stripe Portal
// @description Create a Stripe portal session for managing the subscription and return the session Id to the client. The client can then redirect the user to this session to manage the monetr subscription completely within Stripe.
// @Security ApiKeyAuth
// @Produce json
// @Router /billing/portal [get]
// @Success 200 {array} swag.CreatePortalSessionResponse
// @Failure 402 {object} SubscriptionNotActiveError Returned if the user does not have an active subscription, this endpoint can only be used to update or cancel active subscriptions.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) handleGetStripePortal(ctx iris.Context) {
	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to verify subscription is active")
		return
	}

	if !account.IsSubscriptionActive() {
		c.returnError(ctx, http.StatusPaymentRequired, "subscription is not active")
		return
	}

	returnUrl := ctx.GetReferrer().Raw
	if returnUrl == "" {
		returnUrl = fmt.Sprintf("https://%s", c.configuration.UIDomainName)
	}

	params := &stripe.BillingPortalSessionParams{
		Configuration: nil,
		Customer:      account.StripeCustomerId,
		OnBehalfOf:    nil,
		ReturnURL:     stripe.String(returnUrl),
	}

	session, err := c.stripe.NewPortalSession(c.getContext(ctx), params)
	if err != nil {
		c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create new stripe portal session")
		return
	}

	ctx.JSON(swag.CreatePortalSessionResponse{
		URL: session.URL,
	})
}

// Wait For Stripe Subscription
// @Summary Wait For Stripe Subscription
// @id wait-for-stripe-subscription
// @tags Billing
// @description Long poll endpoint to check to see if the subscription is activated yet. It will return 200 if the subscription is active. Otherwise it will block for 30 seconds and then return 408 if nothing has changed.
// @Security ApiKeyAuth
// @Router /billing/wait [get]
// @Success 200
// @Success 408
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) waitForSubscription(ctx iris.Context) {
	log := c.getLog(ctx)

	repo := c.mustGetAuthenticatedRepository(ctx)
	account, err := repo.GetAccount(c.getContext(ctx))
	if err != nil {
		c.wrapPgError(ctx, err, "failed to retrieve account data")
		return
	}

	if account.IsSubscriptionActive() {
		log.Trace("account has active subscription, nothing to be done")
		return
	}

	channelName := fmt.Sprintf("account:%d:subscription:activated", account.AccountId)
	log = log.WithField("channel", channelName)

	listener, err := c.ps.Subscribe(c.getContext(ctx), channelName)
	if err != nil {
		c.wrapPgError(ctx, err, "failed to listen on channel")
		return
	}
	defer func() {
		if err = listener.Close(); err != nil {
			log.WithError(err).Error("failed to gracefully close listener")
		}
	}()

	log.Debug("waiting for account to be activated channel")

	deadLine := time.NewTimer(30 * time.Second)
	defer deadLine.Stop()

	select {
	case <-deadLine.C:
		ctx.StatusCode(http.StatusRequestTimeout)
		log.Trace("timed out waiting for account/subscription to be setup")
		return
	case <-listener.Channel():
		// Just exit successfully, any message on this channel is considered a success.
		log.Trace("subscription activated successfully")
		return
	}
}
