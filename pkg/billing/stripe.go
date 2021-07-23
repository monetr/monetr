package billing

import (
	"context"
	"encoding/json"
	"github.com/getsentry/sentry-go"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/rest-api/pkg/cache"
	"github.com/monetr/rest-api/pkg/crumbs"
	"github.com/monetr/rest-api/pkg/internal/myownsanity"
	"github.com/monetr/rest-api/pkg/internal/stripe_helper"
	"github.com/monetr/rest-api/pkg/pubsub"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v72"
	"strconv"
	"time"
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

func NewStripeWebhookHandler(log *logrus.Entry, cacheClient cache.Cache, db *pg.DB) StripeWebhookHandler {
	repo := NewAccountRepository(log, cacheClient, db)
	ps := pubsub.NewPostgresPubSub(log, db)
	return &baseStripeWebhookHandler{
		log:                  log,
		repo:                 repo,
		billing:              NewBasicBilling(log, repo, ps),
		billingNotifications: ps,
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

		if err := b.billing.UpdateSubscription(span.Context(),
			subscription.Customer.ID,
			subscription.ID,
			validUntil,
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
			log.WithError(err).Errorf("failed to retrieve account by customer Id")
			return errors.Wrap(err, "failed to retrieve account by customer Id")
		}

		if hub := sentry.GetHubFromContext(span.Context()); hub != nil {
			hub.Scope().SetUser(sentry.User{
				ID: strconv.FormatUint(account.AccountId, 10),
			})
		}

		// Remove the stripe customer Id from the account record.
		account.StripeCustomerId = nil
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
