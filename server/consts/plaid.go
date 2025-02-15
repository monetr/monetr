package consts

import (
	"github.com/plaid/plaid-go/v30/plaid"
)

var (
	PlaidClientName = "monetr"
	PlaidLanguage   = "en"
	PlaidProducts   = []plaid.Products{
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
