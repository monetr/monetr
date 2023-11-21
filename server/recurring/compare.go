package recurring

import (
	"regexp"
	"strings"

	"github.com/adrg/strutil"
)

var (
	cleanStringRegex = regexp.MustCompile(`[a-zA-Z\d]+`)
)

type TransactionNameComparator interface {
	CompareTransactionName(a, b Transaction) float64
}

type TransactionMerchantComparator interface {
	CompareTransactionMerchant(a, b Transaction) float64
}

// sanitizeString takes an input string and removes all non-alphanumeric characters except for underscore and dash.
func sanitizeString(input string) string {
	parts := cleanStringRegex.FindAllString(input, len(input))
	return strings.Join(parts, " ")
}

// equalizeLengths takes two input strings and determines which one is shorter. It then appends a nonsensical character
// to the end of the shorter string to make the two strings equal lengths. This can help make certain text comparison
// algorithms more accurate.
func equalizeLengths(a, b string) (string, string) {
	if len(a) > len(b) {
		b += strings.Repeat("☐", len(a)-len(b))
	} else if len(a) < len(b) {
		a += strings.Repeat("☐", len(b)-len(a))
	}

	return a, b
}

type transactionComparatorBase struct {
	impl            strutil.StringMetric
	equalizeLengths bool
}

func (t *transactionComparatorBase) CompareTransactionName(a, b Transaction) float64 {
	nameA := sanitizeString(a.OriginalName)
	nameB := sanitizeString(b.OriginalName)
	if t.equalizeLengths {
		nameA, nameB = equalizeLengths(nameA, nameB)
	}
	return t.impl.Compare(nameA, nameB)
}

func (t *transactionComparatorBase) CompareTransactionMerchant(a, b Transaction) float64 {
	var merchantA, merchantB string
	if a.OriginalMerchantName != nil {
		merchantA = *a.OriginalMerchantName
	}
	if b.OriginalMerchantName != nil {
		merchantB = *b.OriginalMerchantName
	}
	return t.impl.Compare(merchantA, merchantB)
}
