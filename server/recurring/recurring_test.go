package recurring

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecurringDetection(t *testing.T) {
	data := GetFixtures(t, "monetr_sample_data_1.json")
	// data := GetFixtures(t, "Result_3.json")
	// data := GetFixtures(t, "full sample.json")

	timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
	detection := NewRecurringTransactionDetection(timezone)
	for i := range data {
		detection.AddTransaction(&data[i])
	}

	result := detection.GetRecurringTransactions()
	assert.NotEmpty(t, result)

	j, err := json.MarshalIndent(result, "", "    ")
	require.NoError(t, err, "must be able to marshall result")

	fmt.Println(string(j))
}
