package myownsanity_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/monetr/monetr/server/internal/myownsanity"
	"github.com/stretchr/testify/assert"
)

func TestMapErr(t *testing.T) {
	t.Run("maps every item", func(t *testing.T) {
		result, err := myownsanity.MapErr([]int{1, 2, 3}, func(arg int) (string, error) {
			return strconv.Itoa(arg), nil
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"1", "2", "3"}, result)
	})

	t.Run("stops on the first error", func(t *testing.T) {
		calls := 0
		result, err := myownsanity.MapErr([]int{1, 2, 3}, func(arg int) (string, error) {
			calls++
			if arg == 2 {
				return "", errors.New("bad item")
			}
			return strconv.Itoa(arg), nil
		})
		assert.EqualError(t, err, "bad item")
		assert.Nil(t, result)
		assert.Equal(t, 2, calls, "should not call the callback after an error")
	})

	t.Run("empty input", func(t *testing.T) {
		result, err := myownsanity.MapErr([]int{}, func(arg int) (string, error) {
			return strconv.Itoa(arg), nil
		})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}
