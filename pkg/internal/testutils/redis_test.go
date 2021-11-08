package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRedisPool(t *testing.T) {
	redis := GetRedisPool(t)
	assert.NotNil(t, redis, "resulting redis pool should not be nil")
}
