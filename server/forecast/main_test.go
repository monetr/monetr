package forecast

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed fixtures/*.json
var forecastingFixtureData embed.FS

func TestFixtures(t *testing.T)  {
	files, err := forecastingFixtureData.ReadDir("fixtures")
	assert.NoError(t, err, "must be able to read the fixtures directory")
	assert.NotEmpty(t, files, "must have some fixtures in the fixtures directory")
}
