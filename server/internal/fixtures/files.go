package fixtures

import (
	"embed"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed files/*
var fixtureData embed.FS

func LoadFile(t *testing.T, name string) []byte {
	data, err := fixtureData.ReadFile(path.Join("files", name))
	require.NoError(t, err, "must be able to load fixture data file: %s", name)
	return data
}
