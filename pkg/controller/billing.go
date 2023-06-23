package controller

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/pkg/build"
	"github.com/monetr/monetr/pkg/config"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/stripe_helper"
	"github.com/stripe/stripe-go/v74"
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
	if !c.configuration.Stripe.IsBillingEnabled() {
		return c.notFound(ctx, "billing is not enabled")
	}

	isActive, err := c.paywall.GetSubscriptionIsActive(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to verify there is not already an active subscription")
	}

	// If the customer already has an active subscription we do not want them to try to use this at this time. I don't
	// know how stripe handles this off the top of my head at the time of writing this. But I've setup an endpoint to
	// manage the subscriptions that already exist via the stripe portal. So existing subscriptions should be managed
	// there instead.
	if isActive {
		return c.badRequest(ctx, "there is already an active subscription for your account")
	}

	hasSubscription, err := c.paywall.GetHasSubscription(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to verify that a subscription does not already exist")
	}

	if hasSubscription {
		return c.badRequest(ctx, "there is already a subscription associated with your account")
	}

	var request struct {
		// Specify a specific Stripe Price ID to be used when creating the checkout session. If this is left blank then
		// the default price will be used for the checkout session.
		PriceId *string `json:"priceId"`
		// The path that the user should be returned to if they exit the checkout session.
		CancelPath *string `json:"cancelPath"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
	}

	log := c.getLog(ctx)

	var plan config.Plan
	var priceId string
	if request.PriceId != nil {
		priceId = *request.PriceId
	}

	if priceId == "" && c.configuration.Stripe.InitialPlan != nil {
		priceId = c.configuration.Stripe.InitialPlan.StripePriceId
		plan = *c.configuration.Stripe.InitialPlan
	} else {
		if priceId == "" {
			return c.badRequest(ctx, "must provide a price id")
		}

		{ // Validate the price against our configuration.
			var foundValidPlan bool
			for _, planItem := range c.configuration.Stripe.Plans {
				if planItem.StripePriceId == priceId {
					foundValidPlan = true
					plan = planItem
					break
				}
			}

			if !foundValidPlan {
				return c.badRequest(ctx, "invalid price Id provided")
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
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve account")
	}

	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve current user details")
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
					"accountId":   strconv.FormatUint(me.AccountId, 10),
				},
			},
		})
		if err != nil {
			log.WithError(err).Error("failed to create stripe customer for checkout")
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create stripe customer")
		}

		account.StripeCustomerId = &customer.ID
		if err = c.accounts.UpdateAccount(c.getContext(ctx), account); err != nil {
			log.WithError(err).Error("failed to update account with new customer Id")
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to update account with new customer Id")
		}

		log.Info("successfully created stripe customer for account")
	}

	successUrl := fmt.Sprintf("%s://%s/account/subscribe/after?session={CHECKOUT_SESSION_ID}", c.configuration.ExternalURLProtocol, c.configuration.UIDomainName)
	cancelUrl := fmt.Sprintf("%s://%s/account/subscribe", c.configuration.ExternalURLProtocol, c.configuration.UIDomainName)
	if request.CancelPath != nil {
		cancelUrl = fmt.Sprintf("%s://%s%s", c.configuration.ExternalURLProtocol, c.configuration.UIDomainName, *request.CancelPath)
	}

	crumbs.Debug(c.getContext(ctx), "Creating Stripe Checkout Session", map[string]interface{}{
		"successUrl":   successUrl,
		"cancelUrl":    cancelUrl,
		"collectTaxes": c.configuration.Stripe.TaxesEnabled,
	})

	var params stripe.Params

	// If we are collecting taxes we require the user's billing address. We will not store this information in monetr
	// but Stripe requires it for billing.
	if c.configuration.Stripe.TaxesEnabled {
		params.Extra = &stripe.ExtraValues{
			Values: url.Values{
				"customer_update[address]": []string{
					"auto",
				},
			},
		}
	}

	checkoutParams := &stripe.CheckoutSessionParams{
		Params: params,
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(c.configuration.Stripe.TaxesEnabled),
		},
		AllowPromotionCodes: stripe.Bool(true),
		SuccessURL:          &successUrl,
		CancelURL:           &cancelUrl,
		Customer:            account.StripeCustomerId,
		Discounts:           nil,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				// Number of bank accounts?
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
				"accountId":   strconv.FormatUint(me.AccountId, 10),
			},
			TransferData:    nil,
			TrialEnd:        nil,
			TrialFromPlan:   nil,
			TrialPeriodDays: nil,
		},
	}

	if plan.FreeTrialDays > 0 && account.StripeSubscriptionId == nil {
		checkoutParams.SubscriptionData.TrialPeriodDays = stripe.Int64(int64(plan.FreeTrialDays))
	}

	result, err := c.stripe.NewCheckoutSession(c.getContext(ctx), checkoutParams)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create checkout session")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"sessionId": result.ID,
		"url":       result.URL,
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
	if !c.configuration.Stripe.IsBillingEnabled() {
		return c.notFound(ctx, "billing is not enabled")
	}

	checkoutSessionId := ctx.Param("checkoutSessionId")
	if checkoutSessionId == "" {
		return c.badRequest(ctx, "checkout session Id is required")
	}

	checkoutSession, err := c.stripe.GetCheckoutSession(c.getContext(ctx), checkoutSessionId)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not retrieve Stripe checkout session")
	}

	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve account details")
	}

	stripeCustomerId := ""
	if account.StripeCustomerId != nil {
		stripeCustomerId = *account.StripeCustomerId
	}

	if checkoutSession.Customer.ID != stripeCustomerId {
		crumbs.IndicateBug(c.getContext(ctx), "BUG: The Stripe customer Id for this account does not match the one from the checkout session", map[string]interface{}{
			"accountCustomerId":         account.StripeCustomerId,
			"checkoutSessionCustomerId": checkoutSession.Customer.ID,
		})
	}

	// Now retrieve the subscription status for the latest subscription.
	subscription, err := c.stripe.GetSubscription(c.getContext(ctx), checkoutSession.Subscription.ID)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve subscription details from stripe")
	}

	validUntil := myownsanity.TimeP(time.Unix(subscription.CurrentPeriodEnd, 0))

	if err = c.billing.UpdateCustomerSubscription(
		c.getContext(ctx),
		account,
		subscription.Customer.ID, subscription.ID,
		subscription.Status,
		validUntil,
		time.Now(),
	); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to update subscription state")
	}

	if stripe_helper.SubscriptionIsActive(*subscription) {
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

// Get Stripe Portal
// @id get-stripe-portal
// @tags Billing
// @Summary Get Stripe Portal
// @description Create a Stripe portal session for managing the subscription and return the session Id to the client. The client can then redirect the user to this session to manage the monetr subscription completely within Stripe.
// @Security ApiKeyAuth
// @Produce json
// @Router /billing/portal [get]
// @Success 200 {array} swag.CreatePortalSessionResponse
// @Failure 400 {object} ApiError Returned if the customer does not have a subscription, even if that subscription is expired.
// @Failure 500 {object} ApiError Something went wrong on our end.
func (c *Controller) handleGetStripePortal(ctx echo.Context) error {
	account, err := c.accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to verify subscription is active")
	}

	if !account.HasSubscription() {
		return c.badRequest(ctx, "account does not have a subscription")
	}

	if account.StripeCustomerId == nil {
		crumbs.Debug(c.getContext(ctx), "Account does not have a Stripe customer, a new one will be created.", nil)

		me, err := c.mustGetAuthenticatedRepository(ctx).GetMe(c.getContext(ctx))
		if err != nil {
			crumbs.Error(c.getContext(ctx), "Failed to retrieve the current user to create a Stripe customer", "error", nil)
			return c.wrapPgError(ctx, err, "failed to retrieve current user details")
		}

		name := me.Login.FirstName + " " + me.Login.LastName
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
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve current user details")
		}

		account.StripeCustomerId = &customer.ID

		if err = c.accounts.UpdateAccount(c.getContext(ctx), account); err != nil {
			return c.wrapPgError(ctx, err, "failed to store stripe customer Id")
		}
	}

	// Return the user to the UI home page when they return the monetr. If they are authenticated this will show them
	// the transactions view, or will prompt them for credentials if they are no longer authenticated.
	returnUrl := fmt.Sprintf("%s://%s", c.configuration.ExternalURLProtocol, c.configuration.UIDomainName)
	params := &stripe.BillingPortalSessionParams{
		Configuration: nil,
		Customer:      account.StripeCustomerId,
		OnBehalfOf:    nil,
		ReturnURL:     stripe.String(returnUrl),
	}

	session, err := c.stripe.NewPortalSession(c.getContext(ctx), params)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create new stripe portal session")
	}

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"url": session.URL,
	})
}
