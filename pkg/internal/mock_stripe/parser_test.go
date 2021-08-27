package mock_stripe

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestParseStripeForm(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		queryString := `allow_promotion_codes=true&cancel_url=https%3A%2F%2Fhttp%3A%2F%2Flocalhost%3A1234%2Faccount%2Fsubscribe&customer=cus_KEvUDWuJgVXHZp&line_items[0][price]=price_5xHNDnbnzHO8IGty0n1KCANn&line_items[0][quantity]=1&mode=subscription&payment_method_types[0]=card&subscription_data[metadata][environment]=&subscription_data[metadata][revision]=&subscription_data[metadata][release]=&subscription_data[metadata][accountId]=1087&success_url=https%3A%2F%2Fhttp%3A%2F%2Flocalhost%3A1234%2Faccount%2Fsubscribe%2Fafter%3Fsession%3D%7BCHECKOUT_SESSION_ID%7D`

		values, err := url.ParseQuery(queryString)
		assert.NoError(t, err, "must parse query string successfully")
		assert.NotEmpty(t, values, "values must not be empty")

		result, err := ParseStripeForm(values)
		assert.NoError(t, err, "must parse the stripe form")
		assert.NotEmpty(t, result, "must return at least something")

		assert.Contains(t, result, "line_items", "should contain line items")
		assert.Contains(t, result, "subscription_data", "should contain line items")
		assert.Contains(t, result, "payment_method_types", "should contain line items")
	})
}