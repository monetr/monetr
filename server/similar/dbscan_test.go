package similar

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkPreProcessor(b *testing.B) {
	b.StopTimer()
	fixtureJson := fixtures.LoadFile(b, "monetr_sample_data_1.json")
	var data []models.Transaction
	require.NoError(b, json.Unmarshal(fixtureJson, &data), "must be able to decode fixture data")

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		processor := NewTransactionTFIDF()
		for i := range data {
			processor.AddTransaction(&data[i])
		}

		_ = processor.GetDocuments(context.Background())
	}
}

func BenchmarkDBSCAN(b *testing.B) {
	b.StopTimer()
	fixtureJson := fixtures.LoadFile(b, "monetr_sample_data_1.json")
	var data []models.Transaction
	require.NoError(b, json.Unmarshal(fixtureJson, &data), "must be able to decode fixture data")

	processor := NewTransactionTFIDF()
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	datums := processor.GetDocuments(context.Background())

	dbscan := NewDBSCAN(datums, 0.98, 2)
	ctx := context.Background()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = dbscan.Calculate(ctx)
	}
}

func TestPreProcessor(t *testing.T) {
	fixtureJson := fixtures.LoadFile(t, "monetr_sample_data_1.json")
	var data []models.Transaction
	require.NoError(t, json.Unmarshal(fixtureJson, &data), "must be able to decode fixture data")
	// data := GetFixtures(t, "Result_3.json")
	// data := GetFixtures(t, "full sample.json")
	processor := NewTransactionTFIDF()
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	datums := processor.GetDocuments(context.Background())

	// First test with 0.4 and 3 was excellent!
	// 1.25 is also very good
	dbscan := NewDBSCAN(datums, 0.98, 1)
	result := dbscan.Calculate(context.Background())
	assert.NotEmpty(t, result)
	type Presentation struct {
		ID        models.ID[models.Transaction] `json:"id"`
		Sanitized string                        `json:"sanitized"`
		Original  string                        `json:"original"`
	}
	output := make([][]Presentation, len(result))
	for i, cluster := range result {
		output[i] = make([]Presentation, 0, len(cluster.Items))
		for index := range cluster.Items {
			item := dbscan.dataset[index]
			output[i] = append(output[i], Presentation{
				ID: item.ID,
				// Sanitized: strings.Join(item.UpperParts, " "),
				Original: strings.TrimSpace(item.Transaction.OriginalName + " " + item.Transaction.OriginalMerchantName),
			})
		}
	}

	j, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(j))
}
