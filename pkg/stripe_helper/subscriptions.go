package stripe_helper

import "github.com/stripe/stripe-go/v72"

// SubscriptionIsActive is a helper function that takes in a Stripe subscription object and returns true or false based
// on the state of that object's subscription. This is used to handle scenarios where multiple factors could lead to
// a subscription being active or inactive. At the time of writing this it will return active if the subscription is in
// an active state, or is trialing.
func SubscriptionIsActive(subscription stripe.Subscription) bool {
	switch subscription.Status {
	case stripe.SubscriptionStatusActive, stripe.SubscriptionStatusTrialing:
		return true
	default:
		return false
	}
}
