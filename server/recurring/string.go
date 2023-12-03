package recurring

import (
	"strings"

	"github.com/monetr/monetr/server/models"
)

func CleanNameRegex(transaction *models.Transaction) []string {
	words := clusterCleanStringRegex.FindAllString(
		transaction.OriginalName,
		len(transaction.OriginalName),
	)
	if transaction.OriginalMerchantName != "" {
		words = append(words, clusterCleanStringRegex.FindAllString(
			transaction.OriginalMerchantName,
			len(transaction.OriginalMerchantName),
		)...)
	}
	name := make([]string, 0, len(words))
	for i := range words {
		word := words[i]
		word = strings.ToLower(word)
		word = strings.ReplaceAll(word, "'", "")
		word = strings.ReplaceAll(word, ".", "")
		numbers := numberOnly.FindAllString(word, len(word))
		if len(numbers) > 0 {
			continue
		}
		name = append(name, word)
	}

	return name
}
