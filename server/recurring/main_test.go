package recurring

import (
	"embed"
	"encoding/json"
	"path"
	"testing"

	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/require"
)

//go:embed fixtures/*.json
var fixtureData embed.FS

func GetFixtures(t *testing.T, name string) []models.Transaction {
	data, err := fixtureData.ReadFile(path.Join("fixtures", name))
	require.NoError(t, err, "must be able to load fixture data for recurring transactions")
	var transactions []models.Transaction
	require.NoError(t, json.Unmarshal(data, &transactions), "must be able to decode fixture data")
	return transactions
}
