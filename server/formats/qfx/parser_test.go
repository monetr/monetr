package qfx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("nfcu", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu.qfx")
		items := Parse(string(data))
		assert.NotEmpty(t, items)
		assert.IsType(t, new(Array), items, "Root item should be an array")
	})

	t.Run("us bank", func(t *testing.T) {
		data := GetFixtures(t, "sample-usbank.qfx")
		items := Parse(string(data))
		assert.NotEmpty(t, items)
		assert.IsType(t, new(Array), items, "Root item should be an array")
	})
}
