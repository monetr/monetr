package recurring_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/recurring"
)

func TestTokenize(t *testing.T) {
	t.Run("tokenize long name no merchant", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName:         "WHEN I WORK INC:1233303024 57:COURANT,ELLIOT; 798080132284EPJ. Merchant name: WHEN I WORK INC",
			OriginalMerchantName: "",
		}

		tokens := recurring.Tokenize(&txn)
		j, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(j))
	})

	t.Run("debit card", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS Debit - 1234 - GOOGLE *YOUTUBEPRE G.CO/HELPPAY#CA",
		}

		tokens := recurring.Tokenize(&txn)
		j, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(j))
	})

	t.Run("from manual import", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "ACH Debit - Pwp  Obsidian.md  Privacycom 2111508",
		}

		tokens := recurring.Tokenize(&txn)
		j, _ := json.MarshalIndent(tokens, "", "  ")
		fmt.Println(string(j))
	})
}
