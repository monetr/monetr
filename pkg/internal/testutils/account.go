package testutils

import (
	"github.com/OneOfOne/xxhash"
	"testing"
)

// GetAccountIdForTest is used to create unique accountIds for individual tests. These accountIds are probably unique
// between tests. It is possible for two tests to have the same accountId, but it is very unlikely. This does make sure
// though that a test's accountId does stay the same between test runs as it is not random. This does not guarantee that
// the accountId is present in the database and is primarily aimed at mocks.
func GetAccountIdForTest(t *testing.T) uint64 {
	return xxhash.Checksum64([]byte(t.Name()))
}
