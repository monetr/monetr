package recurring

import (
	"regexp"
	"sort"
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

	nameA = strings.ReplaceAll(nameA, "Merchant name: ", "")
	nameB = strings.ReplaceAll(nameB, "Merchant name: ", "")

	pattern := regexp.MustCompile(`[a-zA-Z\d_-]+`)
	nameAParts := pattern.FindAllString(nameA, len(nameA))
	nameBParts := pattern.FindAllString(nameB, len(nameB))
	sort.Slice(nameAParts, func(i, j int) bool {
		return strings.Compare(strings.ToLower(nameAParts[i]), strings.ToLower(nameAParts[j])) == -1
	})
	sort.Slice(nameBParts, func(i, j int) bool {
		return strings.Compare(strings.ToLower(nameBParts[i]), strings.ToLower(nameBParts[j])) == -1
	})

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
