package recurring

import (
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestTokenizeName(t *testing.T) {
	t.Run("long name no merchant", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName:         "WHEN I WORK INC:1233303024 57:COURANT,ELLIOT; 798080132284EPJ. Merchant name: WHEN I WORK INC",
			OriginalMerchantName: "",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"when", "work", "inc", "courant", "elliot", "merchant", "name", "when", "work", "inc",
		}, lower, "should match the cleaned string")
	})

	t.Run("debit card", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS Debit - 1234 - GOOGLE *YOUTUBEPRE G.CO/HELPPAY#CA",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"google", "youtube premium", "gco", "ca",
		}, lower, "should match the cleaned string")
	})

	t.Run("google cloud", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "Pos Debit 5988 cloud 5zbb G.co/helppay#",
			MerchantName: "Cloud 5zbb",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"cloud", "gco",
		}, lower, "should match the cleaned string")
	})

	t.Run("ach privacy", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "ACH Debit - Pwp Croix Valle Privacycom 2111508",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"croix", "valle",
		}, lower, "should match the cleaned string")
	})

	t.Run("from manual import", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "ACH Debit - Pwp  Obsidian.md  Privacycom 2111508",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"obsidianmd",
		}, lower, "should match the cleaned string")
	})

	t.Run("from manual import two", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS DEBIT-DC    5988 PWP*OBSIDIAN.MD 844-7718229 NY null",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"obsidianmd",
		}, lower, "should match the cleaned string")
	})
}
