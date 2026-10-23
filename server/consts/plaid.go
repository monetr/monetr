package consts

import (
	plaid "github.com/plaid/plaid-go/v49"
)

var (
	PlaidClientName = "monetr"
	PlaidLanguage   = "en"
	PlaidProducts   = []plaid.Products{
		plaid.ProductsTransactions,
	}
)

func PlaidProductStrings() []string {
	items := make([]string, len(PlaidProducts))
	for i, product := range PlaidProducts {
		items[i] = string(product)
	}

	return items
}
