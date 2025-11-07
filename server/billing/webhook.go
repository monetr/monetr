package billing

import (
	"context"
	"encoding/json"
	"time"

	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v81"
)

func (b *baseBilling) HandleStripeWebhook(ctx context.Context, event stripe.Event) error {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"eventType": event.Type,
		"eventId":   event.ID,
	})

	log.Debug("handling webhook from stripe")

	crumbs.Debug(span.Context(), "Handling Stripe webhook.", map[string]any{
		"eventId":  event.ID,
		"liveMode": event.Livemode,
		"type":     event.Type,
	})

	timestamp := time.Unix(event.Created, 0)

	switch event.Type {
	case "checkout.session.completed":
		log.Debugf("checkout session completed")
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		log.Info("handling subscription webhook")
		var subscription stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			log.WithError(err).Errorf("failed to extract subscription from json")
			return errors.Wrap(err, "failed to extract subscription from json")
		}

		validUntil := myownsanity.TimeP(time.Unix(subscription.CurrentPeriodEnd, 0))

		if err := b.UpdateSubscription(
			span.Context(),
			subscription.Customer.ID, subscription.ID,
			subscription.Status,
			validUntil,
			timestamp,
		); err != nil {
			log.WithError(err).Errorf("failed to update subscription")
			return errors.Wrap(err, "failed to update subscription")
		}

		return nil
	case "customer.deleted":
		log.Info("handling customer deleted webhook")
		var customer stripe.Customer
		if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
			log.WithError(err).Errorf("failed to extract customer from json")
			return errors.Wrap(err, "failed to extract customer from json")
		}
		log = log.WithFields(logrus.Fields{
			"stripe": logrus.Fields{
				"customerId": customer.ID,
			},
		})

		account, err := b.accounts.GetAccountByCustomerId(span.Context(), customer.ID)
		if err != nil {
			log.WithError(err).Warn("failed to retrieve account by customer Id")
			crumbs.Warn(span.Context(), "Failed to retrieve an account for this provided customer Id", "stripe", map[string]any{
				"customerId": customer.ID,
			})

			// We don't want this to be treated as an error. There is nothing we can do about it.
			return nil
		}

		log = log.WithFields(logrus.Fields{
			"accountId": account.AccountId,
		})

		crumbs.IncludeUserInScope(span.Context(), account.AccountId)

		// Remove the stripe customer Id from the account record.
		account.StripeCustomerId = nil

		// The subscription would be canceled at this point, and we would need to create a new one. This does mean that
		// the customer would lose access to their invoices and stuff. But this is a result of deleting a customer
		// record entirely.
		account.StripeSubscriptionId = nil
		account.StripeWebhookLatestTimestamp = &timestamp

		if err = b.accounts.UpdateAccount(span.Context(), account); err != nil {
			log.WithError(err).Errorf("failed to remove customer Id from account")
			return errors.Wrap(err, "failed to remove customer Id from account")
		}

		log.Info("removed stripe customer details from account")

		return nil
	default:
		log.Warn("cannot handle stripe webhook event type")
	}

	return nil
}
