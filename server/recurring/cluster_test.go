package recurring

import (
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
}
