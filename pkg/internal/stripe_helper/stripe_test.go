package stripe_helper

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jarcoal/httpmock"
	"github.com/monetr/rest-api/pkg/internal/mock_stripe"
	"github.com/monetr/rest-api/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStripeBase_CreateCustomer(t *testing.T) {
	t.Run("mock success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		mock_stripe.MockStripeGetPriceSuccess(t)

		client := NewStripeHelper(testutils.GetLog(t), gofakeit.UUID())

		prices, err := client.GetPricesById(context.Background(), []string{
			mock_stripe.FakeStripePriceId(t),
			mock_stripe.FakeStripePriceId(t),
			mock_stripe.FakeStripePriceId(t),
		})
		assert.NoError(t, err, "should retrieve prices by id")
		assert.NotNil(t, prices, "should not be nil")
		assert.Len(t, prices, 3, "should have 3 prices")
	})
}

func TestStripeBase_GetPricesById(t *testing.T) {
	t.Run("mock success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		mock_stripe.MockStripeGetPriceSuccess(t)

		client := NewStripeHelper(testutils.GetLog(t), gofakeit.UUID())

		prices, err := client.GetPricesById(context.Background(), []string{
			mock_stripe.FakeStripePriceId(t),
			mock_stripe.FakeStripePriceId(t),
			mock_stripe.FakeStripePriceId(t),
		})
		assert.NoError(t, err, "should retrieve prices by id")
		assert.NotNil(t, prices, "should not be nil")
		assert.Len(t, prices, 3, "should have 3 prices")
	})

	t.Run("mock not found", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		mock_stripe.MockStripeGetPriceNotFound(t)

		client := NewStripeHelper(testutils.GetLog(t), gofakeit.UUID())

		prices, err := client.GetPricesById(context.Background(), []string{
			mock_stripe.FakeStripePriceId(t),
		})
		assert.Error(t, err, "should fail to retrieve price")
		assert.Nil(t, prices, "should be nil")
	})

	t.Run("integration", func(t *testing.T) {
		client := NewStripeHelper(testutils.GetLog(t), GetStripeAPIKeyForTest(t))

		result, err := client.GetPricesById(context.Background(), []string{
			IntegrationTestPriceID,
		})
		assert.NoError(t, err, "request should succeed")
		assert.NotNil(t, result)
		assert.Len(t, result, 1, "should have one price")
	})
}

func TestStripeBase_GetProductsById(t *testing.T) {
	t.Run("mock success", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		mock_stripe.MockStripeListProductsSuccess(t)

		client := NewStripeHelper(testutils.GetLog(t), gofakeit.UUID())

		products, err := client.GetProductsById(context.Background(), []string{
			mock_stripe.FakeStripeProductId(t),
			mock_stripe.FakeStripeProductId(t),
			mock_stripe.FakeStripeProductId(t),
		})
		assert.NoError(t, err, "should retrieve products by id")
		assert.NotNil(t, products, "should not be nil")
		assert.Len(t, products, 3, "should have 3 products")
	})

	t.Run("mock server error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.Deactivate()

		mock_stripe.MockStripeListProductsFailure(t)

		client := NewStripeHelper(testutils.GetLog(t), gofakeit.UUID())

		products, err := client.GetProductsById(context.Background(), []string{
			mock_stripe.FakeStripeProductId(t),
			mock_stripe.FakeStripeProductId(t),
			mock_stripe.FakeStripeProductId(t),
		})
		assert.Error(t, err, "should return an error")
		assert.Nil(t, products, "should not be nil")
	})

	t.Run("integration", func(t *testing.T) {
		client := NewStripeHelper(testutils.GetLog(t), GetStripeAPIKeyForTest(t))

		products, err := client.GetProductsById(context.Background(), []string{
			IntegrationTestProductID,
		})
		assert.NoError(t, err, "should return an error")
		assert.NotNil(t, products, "should not be nil")
		assert.Len(t, products, 1, "should return one product")
	})
}
