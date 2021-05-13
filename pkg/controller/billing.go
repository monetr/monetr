package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/monetrapp/rest-api/pkg/swag"
	"github.com/stripe/stripe-go/v72"
	"net/http"
)

func (c *Controller) handleBilling(p iris.Party) {
	p.Get("/plans", c.getBillingPlans)
	p.Post("/create_checkout", c.handlePostCreateCheckout)
}

func boolP(input bool) *bool {
	return &input
}

func stringP(input string) *string {
	return &input
}

func (c *Controller) getBillingPlans(ctx iris.Context) {

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
