package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRule(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		input := "FREQ=MONTHLY;BYMONTHDAY=15,-1"
		rule, err := NewRule(input)
		assert.NoError(t, err, "must be able to parse semi-monthly rule")
		nextRecurrence := rule.After(time.Date(2022, 4, 5, 0, 0, 0, 0, time.UTC), false).Truncate(24 * time.Hour)
		assert.Equal(t, time.Date(2022, 4, 15, 0, 0, 0, 0, time.UTC), nextRecurrence, "next recurrence should be equal")
	})
}
