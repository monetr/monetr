package recurring

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPreProcessor(t *testing.T) {
	//data := GetFixtures(t, "monetr_sample_data_1.json")
	//data := GetFixtures(t, "Result_3.json")
	data := GetFixtures(t, "full sample.json")
	var processor = &PreProcessor{
		documents: []Document{},
		wc: &WordCount{
			index: 0,
			wc:    map[string][2]int{},
		},
		idf: map[string]float64{},
	}
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	processor.PostPrepareCalculations()

	assert.NotEmpty(t, processor.idf)

	datums := processor.GetDatums()

	// First test with 0.4 and 3 was excellent!
	// 1.25 is also very good
	dbscan := NewDBSCAN(datums, 0.5, 2)
	result := dbscan.Calculate()
	assert.NotEmpty(t, result)
	type Presentation struct {
		ID       uint64    `json:"id"`
		Name     string    `json:"name"`
		Merchant *string   `json:"merchant"`
		Date     time.Time `json:"date"`
		Amount   int64     `json:"amount"`
	}
	output := make([][]Presentation, len(result))
	for i, cluster := range result {
		output[i] = make([]Presentation, 0, len(cluster.Items))
		for _, item := range cluster.Items {
			output[i] = append(output[i], Presentation{
				ID:       item.Transaction.TransactionId,
				Name:     item.Transaction.OriginalName,
				Merchant: item.Transaction.OriginalMerchantName,
				Date:     item.Transaction.Date,
				Amount:   item.Transaction.Amount,
			})
		}
		sort.Slice(output[i], func(x, y int) bool {
			return output[i][x].Date.Before(output[i][y].Date)
		})
	}

	j, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(j))
}

func TestParameters(t *testing.T) {
	//data := GetFixtures(t, "monetr_sample_data_1.json")
	data := GetFixtures(t, "Result_3.json")
	var processor = &PreProcessor{
		documents: []Document{},
		wc: &WordCount{
			index: 0,
			wc:    map[string][2]int{},
		},
		idf: map[string]float64{},
	}
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	processor.PostPrepareCalculations()

	assert.NotEmpty(t, processor.idf)

	datums := processor.GetDatums()

	epsilons := make([]float64, 0)
	for i := 0.1; i < 4.0; i += 0.1 {
		epsilons = append(epsilons, i)
	}
	minPoints := make([]int, 0)
	for i := 1; i < 10; i++ {
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
	var processor = &PreProcessor{
		documents: []Document{},
		wc: &WordCount{
			index: 0,
			wc:    map[string][2]int{},
		},
		idf: map[string]float64{},
	}
	for i := range data {
		processor.AddTransaction(&data[i])
	}

	processor.PostPrepareCalculations()

	assert.NotEmpty(t, processor.idf)

	datums := processor.GetDatums()

	distances := kDistances(datums, 4)
	for _, distance := range distances {
		fmt.Printf("%f\n", distance)
	}
}
