package myownsanity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSliceContains(t *testing.T) {
	data := []string{
		"Item #1",
		"Item #2",
	}

	assert.True(t, SliceContains(data, "Item #1"), "should contain item #1")
	assert.False(t, SliceContains(data, "Item #3"), "should contain item #3")
}

func TestStringPEqual(t *testing.T) {
	{
		a, b := "a", "b"
		assert.False(t, StringPEqual(&a, &b), "should not be equal")
	}

	{
		a, b := "a", "a"
		assert.True(t, StringPEqual(&a, &b), "should be equal")
	}

	{
		a := "a"
		assert.False(t, StringPEqual(&a, nil), "should not be equal")
	}

	{
		b := "b"
		assert.False(t, StringPEqual(nil, &b), "should not be equal")
	}

	assert.True(t, StringPEqual(nil, nil), "should be equal")
}
