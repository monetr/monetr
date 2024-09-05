package controller

import (
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/monetr/monetr/server/billing"
	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/monetr/monetr/server/stripe_helper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v78"
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
	if !c.Configuration.Stripe.IsBillingEnabled() {
		return c.notFound(ctx, "billing is not enabled")
	}

	isActive, err := c.Billing.GetSubscriptionIsActive(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to verify there is not already an active subscription")
	}

	// If the customer already has an active subscription we do not want them to
	// try to use this at this time. I don't know how stripe handles this off the
	// top of my head at the time of writing this. But I've setup an endpoint to
	// manage the subscriptions that already exist via the stripe portal. So
	// existing subscriptions should be managed there instead.
	if isActive {
		return c.badRequest(ctx, "There is already an active subscription for your account")
	}

	hasSubscription, err := c.Billing.GetHasSubscription(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to verify that a subscription does not already exist")
	}

	if hasSubscription {
		return c.badRequest(ctx, "there is already a subscription associated with your account")
	}

	var request struct {
		// The path that the user should be returned to if they exit the checkout
		// session.
		CancelPath *string `json:"cancelPath"`
	}
	if err := ctx.Bind(&request); err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusBadRequest, "malformed JSON")
	}

	log := c.getLog(ctx)
	repo := c.mustGetAuthenticatedRepository(ctx)

	account, err := c.Accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve account")
	}

	me, err := repo.GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve current user details")
	}

	// Check to see if the account does not already have a stripe customer Id. If
	// they don't have one then we want to create one.
	if account.StripeCustomerId == nil {
		log.Debug("account is mmissing a stripe customer ID, one will be created")
		if err := c.Billing.CreateCustomer(c.getContext(ctx), *me.Login, account); err != nil {
			return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to create a Stripe customer")
		}
	}

	successUrl := c.Configuration.Server.GetURL("/account/subscribe/after", map[string]string{
		"session": "{CHECKOUT_SESSION_ID}",
	})
	cancelUrl := c.Configuration.Server.GetURL("/account/subscribe", nil)

	// If a custom cancel path was specified by the requester then use that path.
	// Note: it can only be a path, not a completely custom URL.
	// TODO This still has a code smell to it. If this isn't necessary I think we
	// should just remove it outright.
	if request.CancelPath != nil {
		cancelUrl = c.Configuration.Server.GetURL(*request.CancelPath, nil)
	}

	crumbs.Debug(c.getContext(ctx), "Creating Stripe Checkout Session", map[string]interface{}{
		"successUrl":   successUrl,
		"cancelUrl":    cancelUrl,
		"collectTaxes": c.Configuration.Stripe.TaxesEnabled,
	})

	var params stripe.Params

	// If we are collecting taxes we require the user's billing address. We will
	// not store this information in monetr but Stripe requires it for billing.
	if c.Configuration.Stripe.TaxesEnabled {
		params.Extra = &stripe.ExtraValues{
			Values: url.Values{
				"customer_update[address]": []string{"auto"},
			},
		}
	}

	checkoutParams := &stripe.CheckoutSessionParams{
		Params: params,
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(c.Configuration.Stripe.TaxesEnabled),
		},
		AllowPromotionCodes: stripe.Bool(true),
		SuccessURL:          &successUrl,
		CancelURL:           &cancelUrl,
		Customer:            account.StripeCustomerId,
		Discounts:           nil,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				Price:    &c.Configuration.Stripe.InitialPlan.StripePriceId,
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
		SubmitType:                nil,
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			DefaultTaxRates: nil,
			Metadata: map[string]string{
				"environment": c.Configuration.Environment,
				"revision":    build.Revision,
				"release":     build.Release,
				"accountId":   me.AccountId.String(),
			},
			TransferData: nil,
		},
	}

	result, err := c.Stripe.NewCheckoutSession(c.getContext(ctx), checkoutParams)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to create checkout session")
	}

	log.WithFields(logrus.Fields{
		"stripe": logrus.Fields{
			"sessionId": result.ID,
		},
	}).Debug("created stripe checkout session")

	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"url":       result.URL,
		"sessionId": result.ID,
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
	if !c.Configuration.Stripe.IsBillingEnabled() {
		return c.notFound(ctx, "billing is not enabled")
	}

	checkoutSessionId := ctx.Param("checkoutSessionId")
	if checkoutSessionId == "" {
		return c.badRequest(ctx, "checkout session Id is required")
	}

	checkoutSession, err := c.Stripe.GetCheckoutSession(c.getContext(ctx), checkoutSessionId)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "could not retrieve Stripe checkout session")
	}

	account, err := c.Accounts.GetAccount(c.getContext(ctx), c.mustGetAccountId(ctx))
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
	subscription, err := c.Stripe.GetSubscription(c.getContext(ctx), checkoutSession.Subscription.ID)
	if err != nil {
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "failed to retrieve subscription details from stripe")
	}

	validUntil := myownsanity.TimeP(time.Unix(subscription.CurrentPeriodEnd, 0))

	if err = c.Billing.UpdateCustomerSubscription(
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

func (c *Controller) getBillingPortal(ctx echo.Context) error {
	if !c.Configuration.Stripe.IsBillingEnabled() {
		return c.notFound(ctx, "billing is not enabled")
	}

	me, err := c.mustGetAuthenticatedRepository(ctx).GetMe(c.getContext(ctx))
	if err != nil {
		return c.wrapPgError(ctx, err, "failed to retrieve current user details")
	}

	sessionUrl, err := c.Billing.CreateBillingPortal(
		c.getContext(ctx),
		*me.Login, // Account owner? Assumed?
		c.mustGetAccountId(ctx),
	)

	if err != nil {
		if errors.Cause(err) == billing.ErrMissingSubscription {
			return c.badRequest(ctx, "account does not have a subscription")
		}
		return c.wrapAndReturnError(ctx, err, http.StatusInternalServerError, "Failed to create new stripe portal session")
	}

	return ctx.JSON(http.StatusOK, map[string]any{
		"url": sessionUrl,
	})
}
