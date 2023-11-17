package recurring

import (
	"regexp"
	"strings"

	"github.com/adrg/strutil"
)

type TransactionNameComparator interface {
	CompareTransactionName(a, b Transaction) float64
}

type TransactionMerchantComparator interface {
	CompareTransactionMerchant(a, b Transaction) float64
}

type transactionComparatorBase struct {
	impl strutil.StringMetric
}

func (t *transactionComparatorBase) CompareTransactionName(a, b Transaction) float64 {
	nameA := a.OriginalName
	nameB := b.OriginalName

	pattern := regexp.MustCompile(`[a-zA-Z\d_-]+`)
	nameAParts := pattern.FindAllString(nameA, len(nameA))
	nameBParts := pattern.FindAllString(nameB, len(nameB))

	nameA = strings.Join(nameAParts, " ")
	nameB = strings.Join(nameBParts, " ")

	if len(nameA) > len(nameB) {
		nameB += strings.Repeat("💩", len(nameA)-len(nameB))
	} else if len(nameA) < len(nameB) {
		nameB += strings.Repeat("💩", len(nameB)-len(nameA))
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
