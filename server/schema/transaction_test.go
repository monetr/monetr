package schema_test

import (
	"testing"
	"time"

	"github.com/Oudwins/zog"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/schema"
	"github.com/monetr/monetr/server/util"
	"github.com/stretchr/testify/assert"
)

func TestCreateManualTransaction(t *testing.T) {
	t.Run("happy path, adjust balance", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":           "POS DEBIT COFFEE",
			"amount":         2000,
			"date":           util.Midnight(time.Now(), timezone).UTC().Format(time.RFC3339),
			"adjustsBalance": true,
		}

		var result struct {
			models.Transaction
			AdjustsBalance bool `json:"adjustsBalance"`
		}

		issues := schema.CreateManualTransaction.Merge(schema.AdjustsBalance).Parse(
			input,
			&result,
			zog.WithCtxValue("timezone", timezone),
		)
		assert.Empty(t, zog.Issues.Prettify(issues))
		assert.True(t, result.AdjustsBalance)
	})

	t.Run("happy path, does not adjust balance", func(t *testing.T) {
		timezone := testutils.Must(t, time.LoadLocation, "America/Chicago")
		input := map[string]any{
			"name":           "POS DEBIT COFFEE",
			"amount":         2000,
			"date":           util.Midnight(time.Now(), timezone).UTC().Format(time.RFC3339),
			"adjustsBalance": false,
		}

		var result struct {
			models.Transaction
			AdjustsBalance bool `json:"adjustsBalance"`
		}

		issues := schema.CreateManualTransaction.Merge(schema.AdjustsBalance).Parse(
			input,
			&result,
			zog.WithCtxValue("timezone", timezone),
		)
		assert.Empty(t, zog.Issues.Prettify(issues))
		assert.False(t, result.AdjustsBalance)
	})
}
