package recurring

import "github.com/monetr/monetr/server/models"

type TransactionSearch struct {
	nameComparator     TransactionNameComparator
	merchantComparator TransactionMerchantComparator
}

func (t *TransactionSearch) FindSimilarTransactions(baseline models.Transaction, all []models.Transaction) []models.Transaction {
	result := make([]models.Transaction, 0, len(all))
	for i := range all {
		transaction := all[i]
		if baseline.TransactionId == transaction.TransactionId {
			continue
		}
		name := t.nameComparator.CompareTransactionName(baseline, transaction)
		merchant := t.merchantComparator.CompareTransactionMerchant(baseline, transaction)
		if name > 0.83 || merchant > 0.83 {
			result = append(result, transaction)
		}
	}

	return result
}
