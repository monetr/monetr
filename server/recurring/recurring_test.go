package recurring

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecurringDetection(t *testing.T) {
	t.Run("amazon sample data", func(t *testing.T) {
		data := GetFixtures(t, "amazon_sample_data_1.json")

		result, err := DetectRecurringTransactions(context.Background(), data)
		assert.NoError(t, err)

		j, err := json.MarshalIndent(result, "", "    ")
		require.NoError(t, err, "must be able to marshall result")

		fmt.Println(string(j))
	})

	t.Run("freshbooks sample data", func(t *testing.T) {
		data := GetFixtures(t, "monetr_freshbooks_data_1.json")

		result, err := DetectRecurringTransactions(context.Background(), data)
		assert.NoError(t, err)

		j, err := json.MarshalIndent(result, "", "    ")
		require.NoError(t, err, "must be able to marshall result")

		fmt.Println(string(j))

		assert.EqualValues(t, 30, result.Best.Frequency, "should recurr every 30 days")
	})
}
