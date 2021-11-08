package testutils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAccountIdForTest(t *testing.T) {
	accountIds := map[uint64]struct{}{}
	for i := 0; i < 10; i++ {
		t.Run(fmt.Sprintf("account ID #%d", i+1), func(t *testing.T) {
			accountId := GetAccountIdForTest(t)
			_, exists := accountIds[accountId]
			assert.False(t, exists, "accountId must not have been seen before")
			accountIds[accountId] = struct{}{}
		})
	}
}
