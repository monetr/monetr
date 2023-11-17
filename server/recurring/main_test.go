package recurring

import (
	"encoding/json"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func GetFixtures(t *testing.T, name string) []Transaction {
	data, err := fixtureData.ReadFile(path.Join("fixtures", name))
	require.NoError(t, err, "must be able to load fixture data for recurring transactions")
	var transactions []Transaction
	require.NoError(t, json.Unmarshal(data, &transactions), "must be able to decode fixture data")
	return transactions
}
