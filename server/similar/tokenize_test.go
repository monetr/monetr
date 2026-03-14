package similar_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/similar"
	"github.com/stretchr/testify/assert"
)

func TestTokenize(t *testing.T) {
	t.Run("tokenize long name no merchant", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName:         "WHEN I WORK INC:1233303024 57:COURANT,ELLIOT; 798080132284EPJ. Merchant name: WHEN I WORK INC",
			OriginalMerchantName: "",
		}

		tokens := similar.Tokenize(&txn)
		j, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(j))
	})

	t.Run("debit card", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS Debit - 1234 - GOOGLE *YOUTUBEPRE G.CO/HELPPAY#CA",
		}

		tokens := similar.Tokenize(&txn)
		j, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(j))
	})

	t.Run("from manual import", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "ACH Debit - Pwp  Obsidian.md  Privacycom 2111508",
		}

		tokens := similar.Tokenize(&txn)
		j, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(j))
	})

	t.Run("for blog post", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS DEBIT-DC 5988 TST* CARIBOU COFFE NORTH BRANCH MN",
		}

		tokens := similar.Tokenize(&txn)
		j, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(j))

		strs := make([]string, 0, len(tokens))
		for _, token := range tokens {
			strs = append(strs, token.Final...)
		}

		fmt.Println(strings.Join(strs, " "))

	})
}

func TestCleanNameRegex(t *testing.T) {
	t.Run("long no merchant", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName:         "WHEN I WORK INC:1233303024 57:COURANT,ELLIOT; 798080132284EPJ. Merchant name: WHEN I WORK INC",
			OriginalMerchantName: "",
		}

		lower, _ := similar.CleanNameRegex(&txn)
		assert.EqualValues(t, []string{
			"when", "work", "inc", "courant", "elliot", "merchant", "name", "when", "work", "inc",
		}, lower, "should match the cleaned string")
	})

	t.Run("github", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName:         "GITHUB. Merchant name: GITHUB",
			OriginalMerchantName: "GitHub",
		}

		lower, _ := similar.CleanNameRegex(&txn)
		assert.EqualValues(t, []string{
			"github", "merchant", "name", "github", "github",
		}, lower, "should match the cleaned string")
	})

	t.Run("debit card", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS Debit - 1234 - GOOGLE *YOUTUBEPRE G.CO/HELPPAY#CA",
		}

		lower, _ := similar.CleanNameRegex(&txn)
		assert.EqualValues(t, []string{
			"pos", "debit", "google", "youtubepre", "gco", "helppay", "ca",
		}, lower, "should match the cleaned string")
	})

	t.Run("ach privacy", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "ACH Debit - Pwp Croix Valle Privacycom 2111508",
		}

		lower, _ := similar.CleanNameRegex(&txn)
		assert.EqualValues(t, []string{
			"ach", "debit", "croix", "valle", "privacycom",
		}, lower, "should match the cleaned string")
	})
}

func BenchmarkCleanNameRegex(b *testing.B) {
	b.StopTimer()
	txn := models.Transaction{
		OriginalName:         "WHEN I WORK INC:1233303024 57:COURANT,ELLIOT; 798080132284EPJ. Merchant name: WHEN I WORK INC",
		OriginalMerchantName: "",
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		similar.CleanNameRegex(&txn)
	}
}
