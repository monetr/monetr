package calc_test

import (
	"testing"

	"github.com/monetr/monetr/server/internal/calc"
	"github.com/stretchr/testify/assert"
)

func TestConvertStringToCents(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		type Test struct {
			Input    string
			Expected int64
		}
		tests := []Test{
			{
				Input:    "-10.12",
				Expected: -1012,
			},
			{
				Input:    "12239.99",
				Expected: 1223999,
			},
		}

		for i := range tests {
			test := tests[i]

			result, err := calc.ConvertStringToCents(test.Input)
			assert.NoError(t, err, "should not return an error for these")
			assert.Equal(t, test.Expected, result, "should match the expected output")
		}
	})

	t.Run("failures", func(t *testing.T) {
		type Test struct {
			Input    string
			Expected string
		}
		tests := []Test{
			{
				Input:    "1,230.20",
				Expected: "failed to convert string amount to cents: expected end of string, found ','",
			},
		}

		for i := range tests {
			test := tests[i]

			result, err := calc.ConvertStringToCents(test.Input)
			assert.EqualError(t, err, test.Expected)
			assert.Zero(t, result)
		}
	})
}
