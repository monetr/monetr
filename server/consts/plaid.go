package consts

import (
	"github.com/plaid/plaid-go/v14/plaid"
)

var (
	PlaidClientName = "monetr"
	PlaidLanguage   = "en"
	PlaidCountries  = []plaid.CountryCode{
		plaid.COUNTRYCODE_US,
	}
	PlaidProducts = []plaid.Products{
		plaid.PRODUCTS_TRANSACTIONS,
	}
)

func PlaidProductStrings() []string {
	items := make([]string, len(PlaidProducts))
	for i, product := range PlaidProducts {
		items[i] = string(product)
	}

	return items
}
