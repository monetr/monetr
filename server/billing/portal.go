package billing

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/crumbs"
	. "github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v78"
)

var (
	ErrMissingSubscription = errors.New("account does not have a subscription")
)

// CreateBillingPortal will return a Stripe billing portal URL if the specified
// account ID has a subscription that is "active" and can be updated. If the
// subscription has been cancelled or some other terminal status then the
// portal function will not work. Instead you must create a checkout session.
// Owner represents the Login which "owns" the account and is the one on the
// billing information.
func (b *baseBilling) CreateBillingPortal(
	ctx context.Context,
	owner Login,
	accountId ID[Account],
) (string, error) {
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
		return "", errors.Wrap(err, "cannot determine if account subscription is active")
	}

	span.Status = sentry.SpanStatusOK

	// If the account does not have a subscription, then we cannot create a
	// billing portal for them.
	if !account.HasSubscription() {
		log.Warn("cannot create billing portal, customer is missing a subscription")
		return "", errors.WithStack(ErrMissingSubscription)
	}

	// If the account does not already have a customer ID associated with it, then
	// we need to create one.
	if account.StripeCustomerId == nil {
		log.Debug("account is missing a stripe customer ID, one will be created")
		if err := b.CreateCustomer(span.Context(), owner, account); err != nil {
			return "", err
		}
	}

	// At this point the stripe customer ID should be populated or we should have
	// returned an error. So now we can actually create the billing portal
	// session.
	params := &stripe.BillingPortalSessionParams{
		Configuration: nil,
		Customer:      account.StripeCustomerId,
		Expand:        nil,
		FlowData:      nil,
		Locale:        nil,
		OnBehalfOf:    nil,
		ReturnURL:     stripe.String(b.config.Server.GetBaseURL().String()),
	}

	log.Debug("creating a stripe billing portal")
	session, err := b.stripe.NewPortalSession(span.Context(), params)
	if err != nil {
		return "", err
	}

	return session.URL, nil
}
