package myownsanity_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

func TestIntersection(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		a := []int{1, 2, 3, 4, 5}
		b := []int{2, 4, 1, 8}

		result := myownsanity.Intersection(a, b)
		assert.Equal(t, []int{2, 4, 1}, result)
	})

	t.Run("handles duplicates", func(t *testing.T) {
		a := []int{1, 2, 3, 4, 5, 2, 4, 1}
		b := []int{2, 4, 1, 8, 2, 3}

		result := myownsanity.Intersection(a, b)
		assert.Equal(t, []int{2, 4, 1, 3}, result)
	})
}
