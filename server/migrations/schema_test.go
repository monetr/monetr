package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneral(t *testing.T) {
	files, err := embededMigrations.ReadDir(".")
	assert.NoError(t, err)
	assert.NotEmpty(t, files)
}
