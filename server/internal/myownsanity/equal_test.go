package myownsanity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPEqual(t *testing.T) {
	t.Run("same value different pointers", func(t *testing.T) {
		a, b := "a", "a"
		assert.True(t, PEqual(&a, &b), "should be equal")
	})

	t.Run("different values", func(t *testing.T) {
		a, b := "a", "b"
		assert.False(t, PEqual(&a, &b), "should not be equal")
	})

	t.Run("first is nil", func(t *testing.T) {
		b := "b"
		assert.False(t, PEqual(nil, &b), "should not be equal")
	})

	t.Run("second is nil", func(t *testing.T) {
		a := "a"
		assert.False(t, PEqual(&a, nil), "should not be equal")
	})

	t.Run("both nil", func(t *testing.T) {
		assert.True(t, PEqual[string](nil, nil), "should be equal")
	})

	t.Run("named string type", func(t *testing.T) {
		// Same shape as our ID types, we can't import models here
		type id string
		a, b := id("spend_1"), id("spend_1")
		assert.True(t, PEqual(&a, &b), "should be equal")
	})
}
