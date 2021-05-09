package migrations

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGeneral(t *testing.T) {
	files, err := things.ReadDir(".")
	assert.NoError(t, err)
	assert.NotEmpty(t, files)
}
