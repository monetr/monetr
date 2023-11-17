package recurring

import (
	"testing"
	"time"

	"github.com/adrg/strutil/metrics"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestSearchTransactions(t *testing.T) {
	t.Run("deposit no merchant name", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		data := GetFixtures(t, "monetr_sample_data_1.json")
		comparison := &transactionComparatorBase{
			impl: &metrics.Hamming{
				CaseSensitive: false,
			},
		}
		searcher := &TransactionSearch{
			nameComparator:     comparison,
			merchantComparator: comparison,
		}

		baseline := Transaction{
			TransactionId: 290,
			Amount:        -10000,
			OriginalCategories: []string{
				"Transfer",
				"Credit",
			},
			Date:                 time.Date(2021, 7, 13, 0, 0, 0, 0, timezone),
			OriginalName:         "WHEN I WORK INC:1233303024 57:COURANT,ELLIOT; 798080132284EPJ. Merchant name: WHEN I WORK INC",
			OriginalMerchantName: nil,
		}

		result := searcher.FindSimilarTransactions(baseline, data)
		assert.NotEmpty(t, result, "should have found at least some transactions")
	})
}