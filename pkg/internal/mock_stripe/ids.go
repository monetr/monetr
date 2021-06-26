package mock_stripe

import (
	"fmt"
	"github.com/monetr/rest-api/pkg/internal/testutils"
	"github.com/stretchr/testify/require"
	"testing"
)

func FakeStripePriceId(t *testing.T) string {
	id := fmt.Sprintf("price_%s", testutils.MustGenerateRandomString(t, 24))
	require.NotEmpty(t, id, "stripe price id cannot be empty")
	return id
}

func FakeStripeProductId(t *testing.T) string {
	id := fmt.Sprintf("prod_%s", testutils.MustGenerateRandomString(t, 14))
	require.NotEmpty(t, id, "stripe product id cannot be empty")
	return id
}

func FakeStripePaymentMethodId(t *testing.T) string {
	id := fmt.Sprintf("pm_%s", testutils.MustGenerateRandomString(t, 14))
	require.NotEmpty(t, id, "stripe payment method id cannot be empty")
	return id
}

func FakeStripeCustomerId(t *testing.T) string {
	id := fmt.Sprintf("cus_%s", testutils.MustGenerateRandomString(t, 14))
	require.NotEmpty(t, id, "stripe customer id cannot be empty")
	return id
}
