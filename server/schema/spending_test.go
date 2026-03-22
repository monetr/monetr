package schema_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSpendingSchema(t *testing.T) {
	data, err := json.MarshalIndent(map[string]any{
		"spendingId":        "foo",
		"bankAccountId":     "bac_01gds6eqsq7h5mgevwtmw3cyxb",
		"fundingScheduleId": "fund_01gds6eqsq7h5mgevwtmw3cyxb",
		"name":              "Payday",
		"description":       "15th and the Last day of every month",
		"ruleset":           "DTSTART:20211231T060000Z\nRRULE:FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=15,-1",
		"nextRecurrence":    time.Now(),
		"targetAmount":      1000,
		"excludeWeekends":   true,
	}, "", "  ")
	require.NoError(t, err)
	spending, errs := schema.Parse(
		schema.CreateSpendingSchema,
		bytes.NewBuffer(data),
		models.Spending{},
	)
	fmt.Println()
	assert.Empty(t, errs)
	fmt.Println()
	j, _ := json.MarshalIndent(z.Issues.Treeify(errs), "", "  ")
	fmt.Println(string(j))
	fmt.Println()
	j, _ = json.MarshalIndent(spending, "", "  ")
	fmt.Println(string(j))
}
