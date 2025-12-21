package recurring

import (
	"strings"

	"github.com/monetr/monetr/server/models"
)

var (
	replacementTable = map[string][]string{
		"amz":         {"Amazon"},
		"amzn":        {"Amazon"},
		"amzncom":     {"Amazon"},
		"amazonc":     {"Amazon"},
		"amazoncom":   {"Amazon"},
		"wwwamazonco": {"Amazon"},
		"youtubepre":  {"Youtube", "Premium"},
		"youtubeprem": {"Youtube", "Premium"},
		"coffe":       {"Coffee"},
		"lak":         {"Lake"},
	}
)

type Token struct {
	// Original represents a single word or token in the provided transaction
	// string. This string is unmodified from the input.
	Original string
	// Index represents the position of this token in the original transaction
	// string if we are splitting by whitespace.
	Index int
	// Excluded is true if the token does not include anything significant enough
	// that it should be included in the similarity caluclation
	Excluded bool
	// Equivalent represents the santized version of the same string without
	// modifications made to the case of the string, this may be represented as
	// multiple strings where a single word is being split into two for
	// similarities sake.
	Equivalent []string
	// Final represents the lower sanitized output of the tokenization. This is
	// the value that is used for the final similarity calculation combined with
	// the other tokens from the same transaction.
	Final []string
}

func Tokenize(transaction *models.Transaction) []Token {
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

	tokens := make([]Token, 0, len(words))
	for i := range words {
		word := words[i]
		token := Token{
			Original: word,
			Index:    i,
		}
		equivalent := strings.ReplaceAll(word, "'", "")
		equivalent = strings.ReplaceAll(equivalent, ".", "")

		// Throw out only numeric tokens
		numbers := numberOnly.FindAllString(equivalent, len(equivalent))
		if len(numbers) > 0 {
			token.Excluded = true
			tokens = append(tokens, token)
			continue
		}

		lower := strings.ToLower(equivalent)

		// Throw out tokens that have no vowels
		vowels := vowelsOnly.FindAllString(lower, len(lower))
		if len(vowels) == 0 {
			token.Excluded = true
			tokens = append(tokens, token)
			continue
		}

		// Throw out words that are excluded out right
		if weight, ok := specialWeights[lower]; ok && weight == 0 {
			token.Excluded = true
			tokens = append(tokens, token)
			continue
		}

		// Same with state codes
		if weight, ok := states[lower]; ok && weight == 0 {
			token.Excluded = true
			tokens = append(tokens, token)
			continue
		}

		if replacement, ok := replacementTable[lower]; ok {
			token.Equivalent = replacement
		} else if len(lower) > 2 {
			token.Equivalent = []string{lower}
		} else {
			// Exclude words that are 2 letterss or fewer unless they are a
			// replacement
			token.Excluded = true
			tokens = append(tokens, token)
			continue
		}

		token.Final = make([]string, len(token.Equivalent))
		for i := range token.Equivalent {
			token.Final[i] = strings.ToLower(token.Equivalent[i])
		}

		tokens = append(tokens, token)
	}

	return tokens
}
