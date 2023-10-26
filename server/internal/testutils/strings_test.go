package testutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustGenerateRandomString(t *testing.T) {
	for i := 0; i < 10; i++ {
		length := i + 10
		t.Run(fmt.Sprintf("length %d", length), func(t *testing.T) {
			result := MustGenerateRandomString(t, length)
			assert.NotEmpty(t, result, "random string cannot be empty")
			assert.Len(t, result, length, "must be the desired length")
		})
	}
}
