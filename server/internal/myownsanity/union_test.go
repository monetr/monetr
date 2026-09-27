package myownsanity_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

func TestUnion(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		a := []int{1, 2, 3}
		b := []int{4, 5}

		result := myownsanity.Union(a, b)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
	})

	t.Run("handles duplicates", func(t *testing.T) {
		a := []int{1, 2, 3, 2}
		b := []int{3, 4, 1, 5}

		result := myownsanity.Union(a, b)
		assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
	})

	t.Run("does not modify inputs", func(t *testing.T) {
		backing := []int{1, 2, 3, 99}
		a := backing[:3]
		b := []int{4}

		myownsanity.Union(a, b)
		assert.Equal(t, []int{1, 2, 3, 99}, backing, "spare capacity in a should not be written to")
	})

	t.Run("empty", func(t *testing.T) {
		result := myownsanity.Union([]int{}, []int{})
		assert.Empty(t, result)
	})
}
