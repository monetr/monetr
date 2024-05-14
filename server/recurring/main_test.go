package recurring

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/require"
)

//go:embed fixtures/*.json
var fixtureData embed.FS

func GetFixtures(t *testing.T, name string) []models.Transaction {
	data, err := fixtureData.ReadFile(path.Join("fixtures", name))
	require.NoError(t, err, "must be able to load fixture data for recurring transactions")
	var transactions []models.Transaction
	require.NoError(t, json.Unmarshal(data, &transactions), "must be able to decode fixture data")
	return transactions
}

func TestFixFixtures(t *testing.T) {
	t.Run("monetr_sample_data_1", func(t *testing.T) {
		t.Skip("not needed")
		type Item struct {
			TransactionId        uint64    `json:"transactionId"`
			BankAccountId        uint64    `json:"bankAccountId"`
			Amount               int64     `json:"amount"`
			Date                 time.Time `json:"date"`
			OriginalName         string    `json:"originalName"`
			OriginalMerchantName string    `json:"originalMerchantName"`
		}
		data, err := fixtureData.ReadFile(path.Join("fixtures", "monetr_sample_data_1.json"))
		require.NoError(t, err, "must be able to load fixture data for recurring transactions")
		var transactions []Item
		require.NoError(t, json.Unmarshal(data, &transactions), "must be able to decode fixture data")
		type Updated struct {
			TransactionId        models.ID[models.Transaction] `json:"transactionId"`
			Amount               int64                         `json:"amount"`
			Date                 time.Time                     `json:"date"`
			OriginalName         string                        `json:"originalName"`
			OriginalMerchantName string                        `json:"originalMerchantName"`
		}
		output := make([]Updated, len(transactions))
		for i := range transactions {
			output[i] = Updated{
				TransactionId:        models.NewID(&models.Transaction{}),
				Amount:               transactions[i].Amount,
				Date:                 transactions[i].Date,
				OriginalName:         transactions[i].OriginalName,
				OriginalMerchantName: transactions[i].OriginalMerchantName,
			}
		}

		result, err := json.MarshalIndent(output, "", "  ")
		require.NoError(t, err, "must marshal data back to json properly")
		fmt.Println(string(result))

	})
}
