package ofx

import (
	"embed"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed fixtures/*.qfx
var fixtureData embed.FS

func GetFixtures(t *testing.T, name string) []byte {
	data, err := fixtureData.ReadFile(path.Join("fixtures", name))
	require.NoError(t, err, "must be able to load fixture data for OFX parsing")
	return data
}
