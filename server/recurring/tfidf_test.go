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
			"when", "work", "inc", "courant", "elliot", "when", "work", "inc",
		}, lower, "should match the cleaned string")
	})

	t.Run("debit card", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS Debit - 1234 - GOOGLE *YOUTUBEPRE G.CO/HELPPAY#CA",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"google", "youtube premium", "gco",
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

	t.Run("dominos", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "ACH Debit - Pwp  Domino's 19  Privacycom 2111508",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"dominos",
		}, lower, "should match the cleaned string")
	})

	t.Run("pet supplies", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS DEBIT-DC    5988 CHUCK&DONS FOREST LAKE FOREST LAKE null US",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"chuck", "dons", "forest", "lake", "forest", "lake",
		}, lower, "should match the cleaned string")
	})

	t.Run("market", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS DEBIT-DC    5988 BRINKS MARKET CHISAGO CITY MN US",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"brinks", "market", "chisago", "city",
		}, lower, "should match the cleaned string")
	})

	t.Run("toast pos", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName: "POS DEBIT-DC    5988 TST* CARIBOU COFFE NORTH BRANCH MN",
		}

		lower, _ := TokenizeName(&txn)
		assert.EqualValues(t, []string{
			"caribou", "coffee", "north", "branch",
		}, lower, "should match the cleaned string")
	})
}
