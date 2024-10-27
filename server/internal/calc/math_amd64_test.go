//go:build amd64

// TODO This test will break on arm64
package calc

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCosine(t *testing.T) {
	t.Run("compatability", func(t *testing.T) {
		input := 1.2345
		goImpl := math.Cos(input)
		asmImpl := __cosine64(input)
		// Failing at the moment while I implement it
		assert.InDeltaf(t, goImpl, asmImpl, 1e-6, "should be mostly the same")
	})
}

func TestAbs(t *testing.T) {
	t.Run("compatability", func(t *testing.T) {
		{
			input := -1.2345
			abs := math.Abs(input)
			asm := __absFloat64(input)
			assert.EqualValues(t, abs, asm, "assembly implementation matches builtin")
		}

		{
			input := 1.2345
			abs := math.Abs(input)
			asm := __absFloat64(input)
			assert.EqualValues(t, abs, asm, "assembly implementation matches builtin")
		}
	})
}
