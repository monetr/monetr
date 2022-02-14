package billing

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/pkg/crumbs"
	"github.com/monetr/monetr/pkg/internal/myownsanity"
	"github.com/monetr/monetr/pkg/pubsub"
	"github.com/monetr/monetr/pkg/stripe_helper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
)

type StripeWebhookHandler interface {
	HandleWebhook(ctx context.Context, event stripe.Event) error
}

var (
	_ StripeWebhookHandler = &baseStripeWebhookHandler{}
)

type baseStripeWebhookHandler struct {
	log                  *logrus.Entry
	repo                 AccountRepository
	billing              BasicBilling
	billingNotifications pubsub.PublishSubscribe
}

func NewStripeWebhookHandler(
	log *logrus.Entry,
	accountRepo AccountRepository,
	billing BasicBilling,
	publisher pubsub.PublishSubscribe,
) StripeWebhookHandler {
	return &baseStripeWebhookHandler{
		log:                  log,
		repo:                 accountRepo,
		billing:              billing,
		billingNotifications: publisher,
	}
}

func (b *baseStripeWebhookHandler) HandleWebhook(ctx context.Context, event stripe.Event) error {
	span := sentry.StartSpan(ctx, "Stripe - Handle Webhook")
	defer span.Finish()

	log := b.log.WithContext(span.Context()).WithFields(logrus.Fields{
		"eventType": event.Type,
		"eventId":   event.ID,
	})

	log.Debug("handling webhook from stripe")

	crumbs.Debug(span.Context(), "Handling Stripe webhook.", map[string]interface{}{
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

		var validUntil *time.Time
		if stripe_helper.SubscriptionIsActive(subscription) {
			validUntil = myownsanity.TimeP(time.Unix(subscription.CurrentPeriodEnd, 0))
		}

		if err := b.billing.UpdateSubscription(
			span.Context(),
			subscription.Customer.ID,
			subscription.ID,
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

		account, err := b.repo.GetAccountByCustomerId(span.Context(), customer.ID)
		if err != nil {
			log.WithError(err).Warn("failed to retrieve account by customer Id")
			crumbs.Warn(span.Context(), "Failed to retrieve an account for this provided customer Id", "stripe", map[string]interface{}{
				"customerId": customer.ID,
			})

			// We don't want this to be treated as an error. There is nothing we can do about it.
			return nil
		}

		if hub := sentry.GetHubFromContext(span.Context()); hub != nil {
			hub.Scope().SetUser(sentry.User{
				ID: strconv.FormatUint(account.AccountId, 10),
			})
		}

		// Remove the stripe customer Id from the account record.
		account.StripeCustomerId = nil

		// The subscription would be canceled at this point, and we would need to create a new one. This does mean that
		// the customer would lose access to their invoices and stuff. But this is a result of deleting a customer
		// record entirely.
		account.StripeSubscriptionId = nil
		account.StripeWebhookLatestTimestamp = &timestamp

		if err = b.repo.UpdateAccount(span.Context(), account); err != nil {
			log.WithError(err).Errorf("failed to remove customer Id from account")
			return errors.Wrap(err, "failed to remove customer Id from account")
		}

		return nil
	default:
		log.Warn("cannot handle stripe webhook event type")
	}

	return nil
}
