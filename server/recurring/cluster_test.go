package recurring

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreProcessor(t *testing.T) {
	data := GetFixtures(t, "monetr_sample_data_1.json")
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
	dbscan := NewDBSCAN(datums, 1.25, 3)
	result := dbscan.Calculate()
	assert.NotEmpty(t, result)
	type Presentation struct {
		ID     uint64 `json:"id"`
		Name   string `json:"name"`
		Amount int64  `json:"amount"`
	}
	output := make([][]Presentation, len(result))
	for i, cluster := range result {
		output[i] = make([]Presentation, 0, len(cluster.Items))
		for _, item := range cluster.Items {
			output[i] = append(output[i], Presentation{
				ID:     item.Transaction.TransactionId,
				Name:   item.Transaction.OriginalName,
				Amount: item.Transaction.Amount,
			})
		}
	}

	j, err := json.MarshalIndent(output, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(j))

}
