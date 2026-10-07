package ofx

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizer(t *testing.T) {
	t.Run("nfcu", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu.qfx")
		items, err := Tokenize(t.Context(), data)
		assert.NoError(t, err)
		assert.NotEmpty(t, items)
		assert.IsType(t, new(Array), items, "Root item should be an array")
	})

	t.Run("nfcu wrapped", func(t *testing.T) {
		data := GetFixtures(t, "sample-nfcu-wrapped.qfx")
		items, err := Tokenize(t.Context(), data)
		assert.NoError(t, err)
		assert.NotEmpty(t, items)
		assert.IsType(t, new(Array), items, "Root item should be an array")
	})

	t.Run("us bank", func(t *testing.T) {
		data := GetFixtures(t, "sample-usbank.qfx")
		items, err := Tokenize(t.Context(), data)
		assert.NoError(t, err)
		assert.NotEmpty(t, items)
		assert.IsType(t, new(Array), items, "Root item should be an array")
	})

	t.Run("panics for invalid", func(t *testing.T) {
		data := GetFixtures(t, "invalid.qfx")
		_, err := Tokenize(t.Context(), data)
		assert.Error(t, err)
	})

	t.Run("too deeply nested", func(t *testing.T) {
		data := bytes.Repeat([]byte("<A>"), maxDepth+1)
		_, err := Tokenize(t.Context(), data)
		assert.EqualError(t, err, "OFX data is nested too deep, more than [64] levels at index [64]")
	})

	t.Run("nested to the limit", func(t *testing.T) {
		data := bytes.Repeat([]byte("<A>"), maxDepth)
		items, err := Tokenize(t.Context(), data)
		assert.NoError(t, err, "should be able to tokenize right at the limit")
		assert.IsType(t, new(Array), items, "Root item should be an array")
	})

	t.Run("starts with a closing tag", func(t *testing.T) {
		data := []byte("</A><B>foo")
		_, err := Tokenize(t.Context(), data)
		assert.EqualError(t, err, "syntax error at index [0]")
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		data := GetFixtures(t, "sample-nfcu.qfx")
		_, err := Tokenize(ctx, data)
		assert.EqualError(t, err, "failed to tokenize OFX data: context canceled")
	})
}
