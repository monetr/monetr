package id_test

import (
	"regexp"
	"testing"

	"github.com/monetr/monetr/server/id"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		id := id.New()
		assert.NotEmpty(t, id, "ID string must not be empty")
		assert.Regexp(t, regexp.MustCompile(`^[0-7][0-9a-hjkmnp-tv-z]{25}$`), id, "ID must match the expected pattern and exact length")
	})
}
