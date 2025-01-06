package recurring

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkPreProcessor(b *testing.B) {
	b.StopTimer()
	fixtureJson, err := fixtureData.ReadFile(path.Join("fixtures", "monetr_sample_data_1.json"))
	require.NoError(b, err, "must be able to load fixture data for recurring transactions")
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
	fixtureJson, err := fixtureData.ReadFile(path.Join("fixtures", "monetr_sample_data_1.json"))
	require.NoError(b, err, "must be able to load fixture data for recurring transactions")
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
	data := GetFixtures(t, "monetr_sample_data_1.json")
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
				ID:        item.ID,
				Sanitized: strings.Join(item.Parts, " "),
				Original:  strings.TrimSpace(item.Transaction.OriginalName + " " + item.Transaction.OriginalMerchantName),
			})
		}
	}

	j, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(j))
}

func TestParameters(t *testing.T) {
	if testing.Short() {
		t.Skipf("parameters testing will be skippped for short")
	}
	data := GetFixtures(t, "monetr_sample_data_1.json")
	//data := GetFixtures(t, "Result_3.json")
	processor := NewTransactionTFIDF()
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	datums := processor.GetDocuments(context.Background())

	epsilons := make([]float32, 0)
	for i := float32(0.1); i < 2.0; i += 0.1 {
		epsilons = append(epsilons, i)
	}
	minPoints := make([]int, 0)
	for i := 1; i < 5; i++ {
		minPoints = append(minPoints, i)
	}

	for _, epsilon := range epsilons {
		for _, minPoint := range minPoints {
			dbscan := NewDBSCAN(datums, 0.39, 2)
			result := dbscan.Calculate(context.Background())
			assert.NotEmpty(t, result)
			avgItemsPerCluster := 0
			for _, cluster := range result {
				avgItemsPerCluster += len(cluster.Items)
			}
			fmt.Printf("Epsilon: %f    Min Points: %d    Number of Clusters: %d    Avg Count: %d\n", epsilon, minPoint, len(result), avgItemsPerCluster/len(result))
		}
	}

}
