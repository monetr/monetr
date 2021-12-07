package stripe_helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v72"
)

func TestSubscriptionIsActive(t *testing.T) {
	t.Run("active", func(t *testing.T) {
		subscriptions := []stripe.Subscription{
			{
				Status: stripe.SubscriptionStatusActive,
			},
			{
				Status: stripe.SubscriptionStatusTrialing,
			},
		}

		for _, subscription := range subscriptions {
			assert.True(t, SubscriptionIsActive(subscription), "subscription should be active")
		}
	})

	t.Run("inactive", func(t *testing.T) {
		subscriptions := []stripe.Subscription{
			{
				Status: stripe.SubscriptionStatusCanceled,
			},
			{
				Status: stripe.SubscriptionStatusPastDue,
			},
		}

		for _, subscription := range subscriptions {
			assert.False(t, SubscriptionIsActive(subscription), "subscription should not be active")
		}
	})
}
