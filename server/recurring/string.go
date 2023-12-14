package recurring

import (
	"strings"

	"github.com/monetr/monetr/server/models"
)

func CleanNameRegex(transaction *models.Transaction) (lower []string, normal []string) {
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
	lower = make([]string, 0, len(words))
	normal = make([]string, 0, len(words))
	for i := range words {
		word := words[i]
		word = strings.ReplaceAll(word, "'", "")
		word = strings.ReplaceAll(word, ".", "")
		numbers := numberOnly.FindAllString(word, len(word))
		if len(numbers) > 0 {
			continue
		}
		lower = append(lower, strings.ToLower(word))
		normal = append(normal, word)
	}

	return lower, normal
}

// dedupeStringSlice will remove duplicate words without case sensitivity from the provided string slice while
// preserving the order of the slice.
func dedupeStringSlice(input []string) []string {
	result := make([]string, 0)
	dupe := map[string]struct{}{}
	for _, item := range input {
		lower := strings.ToLower(item)
		if _, ok := dupe[lower]; ok {
			continue
		}

		result = append(result, item)
		dupe[lower] = struct{}{}
	}

	return result
}
