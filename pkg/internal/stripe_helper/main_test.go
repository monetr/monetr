package stripe_helper

import (
	"os"
	"testing"
)

const (
	// These Ids are in our stripe test account. They will always be available and they are only present for testing.
	IntegrationTestPriceID   = "price_1IqmMLI4uGGnwpgwxTW4k2ev"
	IntegrationTestProductID = "prod_JTjqGdmUROFpcq"
)

func GetStripeAPIKeyForTest(t *testing.T) string {
	apiKey := os.Getenv("STRIPE_INTEGRATION_TEST_KEY")
	if apiKey == "" {
		t.Skipf("STRIPE_INTEGRATION_TEST_KEY environment variable is missing")
	}

	return apiKey
}
