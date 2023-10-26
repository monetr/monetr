package mock_stripe

import (
	"fmt"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v72"
	"net/http"
	"testing"
)

func TestMockStripeGetCompletedCheckoutSession(t *testing.T) {
	t.Run("valid id", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock := NewMockStripeHelper(t)

		mock.MockGetCheckoutSession(t)

		var checkoutSession stripe.CheckoutSession
		mock.CreateCheckoutSession(t, &checkoutSession)

		response, err := http.Get(Path(t, fmt.Sprintf("/v1/checkout/sessions/%s", checkoutSession.ID)))
		assert.NoError(t, err, "must be able to create request")
		assert.Equal(t, http.StatusOK, response.StatusCode, "must respond with ok")
	})

	t.Run("bad id", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock := NewMockStripeHelper(t)

		mock.MockGetCheckoutSession(t)

		response, err := http.Get(Path(t, "/v1/checkout/sessions/"))
		assert.NoError(t, err, "must be able to create request")
		assert.Equal(t, http.StatusNotFound, response.StatusCode, "must respond with ok")
	})

	t.Run("not found", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		mock := NewMockStripeHelper(t)

		mock.MockGetCheckoutSession(t)

		response, err := http.Get(Path(t, "/v1/checkout/sessions/not_found"))
		assert.NoError(t, err, "must be able to create request")
		assert.Equal(t, http.StatusNotFound, response.StatusCode, "must respond with ok")
	})
}
