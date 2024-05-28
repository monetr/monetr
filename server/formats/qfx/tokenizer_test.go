package qfx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizer(t *testing.T) {
	t.Run("nfcu", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu.qfx")
		items := Tokenize(string(data))
		assert.NotEmpty(t, items)
		assert.IsType(t, new(Array), items, "Root item should be an array")
	})

	t.Run("nfcu 2", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu-2.qfx")
		items := Tokenize(string(data))
		assert.NotEmpty(t, items)
		assert.IsType(t, new(Field), items, "Root item should be a field for nfcu 2")
	})

	t.Run("us bank", func(t *testing.T) {
		data := GetFixtures(t, "sample-usbank.qfx")
		items := Tokenize(string(data))
		assert.NotEmpty(t, items)
		assert.IsType(t, new(Array), items, "Root item should be an array")
	})

	t.Run("panics for invalid", func(t *testing.T) {
		data := GetFixtures(t, "invalid.qfx")
		assert.PanicsWithValue(t, "QFX file provided is not valid", func() {
			_ = Tokenize(string(data))
		})
	})
}
