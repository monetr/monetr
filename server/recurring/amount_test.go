package recurring

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

func TestAmountProcessor(t *testing.T) {
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

	dbscan := NewDBSCAN(processor.GetDatums(), 0.98, 1)
	result := dbscan.Calculate()
	assert.NotEmpty(t, result)

	for _, cluster := range result {
		transactions := make([]*models.Transaction, 0, len(cluster.Items))
		for index := range cluster.Items {
			transactions = append(transactions, dbscan.dataset[index].Transaction)
		}

		amountProcessor := &AmountPreProcessor{}
		for _, transaction := range transactions {
			amountProcessor.AddTransaction(transaction)
		}

		amountDbscan := NewDBSCAN(amountProcessor.GetDatums(), 0.03, 1)
		amountResult := amountDbscan.Calculate()

		assert.NotEmpty(t, amountResult)
		type Presentation struct {
			ID     uint64    `json:"id"`
			Name   string    `json:"name"`
			Amount int64     `json:"amount"`
			Vector []float64 `json:"vector"`
		}
		output := make([][]Presentation, len(amountResult))
		for i, amountCluster := range amountResult {
			output[i] = make([]Presentation, 0, len(amountCluster.Items))
			for index := range amountCluster.Items {
				item := amountDbscan.dataset[index]
				output[i] = append(output[i], Presentation{
					ID:     item.ID,
					Name:   item.Transaction.OriginalName,
					Amount: item.Transaction.Amount,
					Vector: item.Vector,
				})
			}
		}

		j, err := json.MarshalIndent(output, "", "    ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(j))
	}
}
