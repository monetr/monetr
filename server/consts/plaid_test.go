package consts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaidProductStrings(t *testing.T) {
	assert.NotEmpty(t, PlaidProductStrings(), "must return at least one product string no matter what")
}
