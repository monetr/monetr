package recurring

import (
	"encoding/json"
	"fmt"
	"math"
	"path"
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
		var processor = &PreProcessor{
			documents: []Document{},
			wc:        map[string]int{},
			idf:       map[string]float64{},
		}
		for i := range data {
			processor.AddTransaction(&data[i])
		}

		processor.PostPrepareCalculations()
		_ = processor.GetDatums()
	}
}

func BenchmarkDBSCAN(b *testing.B) {
	b.StopTimer()
	fixtureJson, err := fixtureData.ReadFile(path.Join("fixtures", "monetr_sample_data_1.json"))
	require.NoError(b, err, "must be able to load fixture data for recurring transactions")
	var data []models.Transaction
	require.NoError(b, json.Unmarshal(fixtureJson, &data), "must be able to decode fixture data")

	var processor = &PreProcessor{
		documents: []Document{},
		wc:        map[string]int{},
		idf:       map[string]float64{},
	}
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	processor.PostPrepareCalculations()
	datums := processor.GetDatums()

	dbscan := NewDBSCAN(datums, 0.98, 2)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = dbscan.Calculate()
	}
}

func TestPreProcessor(t *testing.T) {
	data := GetFixtures(t, "monetr_sample_data_1.json")
	//data := GetFixtures(t, "Result_3.json")
	// data := GetFixtures(t, "full sample.json")
	var processor = &PreProcessor{
		documents: []Document{},
		wc:        map[string]int{},
		idf:       map[string]float64{},
	}
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	processor.PostPrepareCalculations()

	assert.NotEmpty(t, processor.idf)

	datums := processor.GetDatums()

	// First test with 0.4 and 3 was excellent!
	// 1.25 is also very good
	dbscan := NewDBSCAN(datums, 0.98, 1)
	result := dbscan.Calculate()
	assert.NotEmpty(t, result)
	type Presentation struct {
		ID        uint64 `json:"id"`
		Sanitized string `json:"sanitized"`
	}
	output := make([][]Presentation, len(result))
	for i, cluster := range result {
		output[i] = make([]Presentation, 0, len(cluster.Items))
		for index := range cluster.Items {
			item := dbscan.dataset[index]
			output[i] = append(output[i], Presentation{
				ID:        item.ID,
				Sanitized: item.String,
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
	var processor = &PreProcessor{
		documents: []Document{},
		wc:        map[string]int{},
		idf:       map[string]float64{},
	}
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	processor.PostPrepareCalculations()

	assert.NotEmpty(t, processor.idf)

	datums := processor.GetDatums()

	epsilons := make([]float64, 0)
	for i := 0.1; i < 2.0; i += 0.1 {
		epsilons = append(epsilons, i)
	}
	minPoints := make([]int, 0)
	for i := 1; i < 5; i++ {
		minPoints = append(minPoints, i)
	}

	for _, epsilon := range epsilons {
		for _, minPoint := range minPoints {
			dbscan := NewDBSCAN(datums, 0.39, 2)
			result := dbscan.Calculate()
			assert.NotEmpty(t, result)
			avgItemsPerCluster := 0
			for _, cluster := range result {
				avgItemsPerCluster += len(cluster.Items)
			}
			fmt.Printf("Epsilon: %f    Min Points: %d    Number of Clusters: %d    Avg Count: %d\n", epsilon, minPoint, len(result), avgItemsPerCluster/len(result))
		}
	}

}

func TestKDistances(t *testing.T) {
	data := GetFixtures(t, "monetr_sample_data_1.json")
	//data := GetFixtures(t, "Result_3.json")
	//data := GetFixtures(t, "full sample.json")
	var processor = &PreProcessor{
		documents: []Document{},
		wc:        map[string]int{},
		idf:       map[string]float64{},
	}
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	processor.PostPrepareCalculations()

	assert.NotEmpty(t, processor.idf)

	datums := processor.GetDatums()

	distances := kDistances(datums, 2)
	distancesFiltered := make([]float64, 0, len(distances))
	for _, distance := range distances {
		if distance < 0.0000001 || math.IsNaN(distance) {
			continue
		}
		distancesFiltered = append(distancesFiltered, distance)
	}
	rates := rollingRateOfChange(1, distancesFiltered)
	rates2 := rollingRateOfChange(1, rates)
	// Log the rates, the rate2.0 will spike when we have a decent epsilon.
	for i, distance := range distancesFiltered {
		fmt.Printf("[%d] %f rate: %f rate2.0: %f\n", i, distance, rates[i], rates2[i])
	}
	//// Find the first big rate of change of rate of change spike. The distance _after_ this will serve as a reasonable
	//// epsilon. Might be more reliable if this was normalized with log()
	//for i := range distancesFiltered {
	//	if rates2[i] > 10000 {
	//		fmt.Println("found epsilon:", distancesFiltered[i+1])
	//		break
	//	}
	//}
}

func rollingRateOfChange(n int, vector []float64) []float64 {
	length := len(vector)
	rates := make([]float64, length)

	for i := n; i < length; i++ {
		// This is just wrong? idk what i was thinking
		previous := vector[i-n]
		current := vector[i]

		if previous != 0 {
			rate := (current - previous) / previous
			rates[i] = rate
		} else {
			// Handle division by zero if needed
			rates[i] = 0
		}
	}

	return rates
}
