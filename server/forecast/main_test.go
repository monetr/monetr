package forecast

import (
	"embed"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/stretchr/testify/assert"
)

//go:embed fixtures/*.json
var forecastingFixtureData embed.FS

func TestFixtures(t *testing.T) {
	files, err := forecastingFixtureData.ReadDir("fixtures")
	assert.NoError(t, err, "must be able to read the fixtures directory")
	assert.NotEmpty(t, files, "must have some fixtures in the fixtures directory")
}

func TestFixFixtures(t *testing.T) {
	spending := make([]map[string]interface{}, 0)
	spendingJson := testutils.Must(t, forecastingFixtureData.ReadFile, "fixtures/elliots-spending-20230705.json")
	testutils.MustUnmarshalJSON(t, spendingJson, &spending)
	assert.NotEmpty(t, spending, "must have spending data loaded")
	for i := range spending {
		spending[i]["spendingId"] = models.NewID(&models.Spending{})
	}
	j, err := json.MarshalIndent(spending, "", "  ")
	assert.NoError(t, err, "must not error")
	fmt.Println(string(j))
}
