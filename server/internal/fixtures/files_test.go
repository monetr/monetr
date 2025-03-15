package fixtures_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestLoadFile(t *testing.T) {
	t.Run("valid fixture file", func(t *testing.T) {
		data := fixtures.LoadFile(t, "sample-part-one.ofx")
		assert.NotEmpty(t, data, "must be able to read sample part one ofx file")

		data = fixtures.LoadFile(t, "sample-part-two.ofx")
		assert.NotEmpty(t, data, "must be able to read sample part two ofx file")
	})
}
