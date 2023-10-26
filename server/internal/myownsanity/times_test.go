package myownsanity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeP(t *testing.T) {
	input := time.Now()
	result := TimeP(input)
	assert.NotNil(t, result, "output must not be nil")
	assert.EqualValues(t, input, *result, "output's value must match input")
}

func TestTimesPEqual(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		input := time.Now()
		a, b := &input, &input
		assert.True(t, TimesPEqual(a, b), "should be equal")
	})

	t.Run("nils", func(t *testing.T) {
		input := time.Now()
		assert.False(t, TimesPEqual(nil, &input), "should not be equal when one is nil")
		assert.False(t, TimesPEqual(&input, nil), "should not be equal when one is nil")
		assert.True(t, TimesPEqual(nil, nil), "but should be equal if they are both nil")
	})
}
