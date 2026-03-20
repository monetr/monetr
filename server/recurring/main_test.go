package recurring

import (
	"encoding/json"
	"testing"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/require"
)

func GetFixtures(t *testing.T, name string) []models.Transaction {
	var data []models.Transaction
	require.NoError(t, json.Unmarshal(fixtures.LoadFile(t, name), &data))
	return data
}
