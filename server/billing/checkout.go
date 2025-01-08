package billing

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/build"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v81"
)

var (
	ErrSubscriptionAlreadyActive = errors.New("There is already an active subscription for your account")
	ErrSubscriptionAlreadyExists = errors.New("There is already a subscription for your account")
)

// CreateCheckout takes a Login object that represents the "billing owner" as
// well as an account ID and an optional cancel path. If the specified account
// has a subscription at all then this function will return an error. If the
// subscription is active it will return an `ErrSubscriptionAlreadyActive` and
// if the subscription is not active but exists it will return
// `ErrSubscriptionAlreadyExists`. Both of these errors should be treated as an
// indication that the user cannot use the checkout to establish billing.
// Instead a billing portal should be created. If the optional cancel path is
// specified then that will be used for the non-success return route from the
// checkout. The default is `/account/subscribe` if one is not provided. The
// success path of the checkout will always be `/account/subscribe/after` and
// cannot be overwritten.
func (b *baseBilling) CreateCheckout(
	ctx context.Context,
	owner Login,
	accountId ID[Account],
	cancelPath *string,
) (*stripe.CheckoutSession, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"loginId":   owner.LoginId,
		"accountId": accountId,
	})

	// Gather the account details from the repo. This data might be cached but
	// should be considered accurate as all writes for subscription data go
	// through this interface.
	account, err := b.accounts.GetAccount(span.Context(), accountId)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "cannot determine if account subscription is active")
	}

	if account.IsSubscriptionActive(b.clock.Now()) && !account.IsTrialing(b.clock.Now()) {
		return nil, errors.WithStack(ErrSubscriptionAlreadyActive)
	}

	if account.HasSubscription() {
		// Even if the subscription isn't active, if they have a subscription then
		// they need to use the billing portal instead.
		return nil, errors.WithStack(ErrSubscriptionAlreadyExists)
	}

	// If the account does not already have a customer ID associated with it, then
	// we need to create one.
	if account.StripeCustomerId == nil {
		log.Debug("account is missing a stripe customer ID, one will be created")
		if err := b.CreateCustomer(span.Context(), owner, account); err != nil {
			return nil, err
		}
	}

	// Success url is the URL they'll be brought to after they complete the
	// checkout. Don't URL encode the checkout session ID because stripe needs it
	// directly.
	successUrl := fmt.Sprintf(
		"%s?session={CHECKOUT_SESSION_ID}",
		b.config.Server.GetURL("/account/subscribe/after", nil),
	)
	// Cancel URL is where they will be brought if they chose to not complete the
	// checkout.
	cancelUrl := b.config.Server.GetURL("/account/subscribe", nil)

	// If a custom cancel path was provided then use that instead.
	if cancelPath != nil {
		cancelUrl = b.config.Server.GetURL(*cancelPath, nil)
	}

	taxesEnabled := b.config.Stripe.TaxesEnabled

	crumbs.Debug(span.Context(), "Creating Stripe Checkout Session", map[string]interface{}{
		"successUrl":   successUrl,
		"cancelUrl":    cancelUrl,
		"collectTaxes": taxesEnabled,
	})
	log.WithFields(logrus.Fields{
		"stripe": logrus.Fields{
			"successUrl":   successUrl,
			"cancelUrl":    cancelUrl,
			"collectTaxes": taxesEnabled,
		},
	}).Debug("creating stripe checkout session")

	var params stripe.Params

	// If we are collecting taxes we require the user's billing address. We will
	// not store this information in monetr but Stripe requires it for billing.
	if taxesEnabled {
		params.Extra = &stripe.ExtraValues{
			Values: url.Values{
				"customer_update[address]": []string{"auto"},
			},
		}
	}

	result, err := b.stripe.NewCheckoutSession(span.Context(), &stripe.CheckoutSessionParams{
		Params: params,
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(taxesEnabled),
		},
		AllowPromotionCodes: stripe.Bool(true),
		SuccessURL:          &successUrl,
		CancelURL:           &cancelUrl,
		Customer:            account.StripeCustomerId,
		Discounts:           nil,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				Price:    &b.config.Stripe.InitialPlan.StripePriceId,
			},
		},
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			DefaultTaxRates: nil,
			Metadata: map[string]string{
				"environment": b.config.Environment,
				"revision":    build.Revision,
				"release":     build.Release,
				"accountId":   accountId.String(),
			},
			TransferData: nil,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create checkout session")
	}

	log.WithFields(logrus.Fields{
		"stripe": logrus.Fields{
			"sessionId": result.ID,
		},
	}).Debug("created stripe checkout session")

	return result, nil
}

// AfterCheckout is called when the user is directed back to the application
// after completing a checkout in stripe. This function takes the checkout
// session ID that is returned during the success redirect from stripe. This
// function then checks to see if the user completed their subscription as part
// of the checkout and takes their subscription status and stores it on the
// account. It then returns true or false indicating whether the subscription is
// now active.
func (b *baseBilling) AfterCheckout(
	ctx context.Context,
	accountId ID[Account],
	checkoutSessionId string,
) (active bool, _ error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"accountId": accountId,
		"stripe": logrus.Fields{
			"checkoutSessionId": checkoutSessionId,
		},
	})

	log.Debug("retrieving checkout session details")
	checkoutSession, err := b.stripe.GetCheckoutSession(
		span.Context(),
		checkoutSessionId,
	)
	if err != nil {
		return false, errors.Wrap(err, "failed to retrieve checkout session")
	}

	// Gather the account details from the repo. This data might be cached but
	// should be considered accurate as all writes for subscription data go
	// through this interface.
	account, err := b.accounts.GetAccount(span.Context(), accountId)
	if err != nil {
		span.Status = sentry.SpanStatusInternalError
		return false, errors.Wrap(err, "could not retrieve account details")
	}

	// Pull the stripe customer ID off of the account. This is to prevent the
	// possiblity that the auth changed during the checkout somehow? We don't want
	// to allow someone to setup a subscription for someone else. This is a huge
	// edge case but makes sure that everything is correct.
	stripeCustomerId := ""
	if account.StripeCustomerId != nil {
		stripeCustomerId = *account.StripeCustomerId
	}

	if checkoutSession.Customer.ID != stripeCustomerId {
		log.WithFields(logrus.Fields{
			"bug": true,
			"stripe": logrus.Fields{
				"checkoutSessionId": checkoutSessionId,
				"customerId": logrus.Fields{
					"account":  stripeCustomerId,
					"checkout": checkoutSession.Customer.ID,
				},
			},
		}).Warn("stripe customer ID from checkout session does not match the stripe customer ID on the account")
		crumbs.IndicateBug(span.Context(), "BUG: The Stripe customer Id for this account does not match the one from the checkout session", map[string]interface{}{
			"accountCustomerId":         account.StripeCustomerId,
			"checkoutSessionCustomerId": checkoutSession.Customer.ID,
		})
		return false, errors.New("stripe customer ID mismatch")
	}

	log = log.WithFields(logrus.Fields{
		"stripe": logrus.Fields{
			"checkoutSessionId": checkoutSessionId,
			"customerId":        stripeCustomerId,
			"subscriptionId":    checkoutSession.Subscription.ID,
		},
	})

	log.Debug("retreiving subscription details from checkout session")
	subscription, err := b.stripe.GetSubscription(
		span.Context(),
		checkoutSession.Subscription.ID,
	)
	if err != nil {
		return false, errors.Wrap(err, "failed to retrieve subscription from checkout session")
	}

	validUntil := myownsanity.TimeP(time.Unix(subscription.CurrentPeriodEnd, 0))
	if err := b.UpdateCustomerSubscription(
		span.Context(),
		account,
		subscription.Customer.ID,
		subscription.ID,
		subscription.Status,
		validUntil,
		b.clock.Now(),
	); err != nil {
		return false, errors.Wrap(err, "failed to update subscription state from checkout session")
	}

	return SubscriptionIsActive(*subscription), nil
}
