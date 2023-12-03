package recurring

import (
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestCleanNameRegex(t *testing.T) {
	t.Run("long no merchant", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName:         "WHEN I WORK INC:1233303024 57:COURANT,ELLIOT; 798080132284EPJ. Merchant name: WHEN I WORK INC",
			OriginalMerchantName: "",
		}

		result := CleanNameRegex(&txn)
		assert.EqualValues(t, []string{
			"when", "i", "work", "inc", "courant", "elliot", "798080132284epj", "merchant", "name", "when", "i", "work", "inc",
		}, result, "should match the cleaned string")
	})

	t.Run("github", func(t *testing.T) {
		txn := models.Transaction{
			OriginalName:         "GITHUB. Merchant name: GITHUB",
			OriginalMerchantName: "GitHub",
		}

		result := CleanNameRegex(&txn)
		assert.EqualValues(t, []string{
			"github", "merchant", "name", "github", "github",
		}, result, "should match the cleaned string")
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
		CleanNameRegex(&txn)
	}
}
