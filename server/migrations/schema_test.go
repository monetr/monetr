package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneral(t *testing.T) {
	files, err := things.ReadDir(".")
	assert.NoError(t, err)
	assert.NotEmpty(t, files)
}
