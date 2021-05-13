package mock_stripe

import (
	"fmt"
	"github.com/monetrapp/rest-api/pkg/internal/testutils"
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
