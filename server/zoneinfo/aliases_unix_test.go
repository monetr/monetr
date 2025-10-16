//go:build unix

package zoneinfo_test

import (
	"os"
	"testing"

	"github.com/monetr/monetr/server/zoneinfo"
	"github.com/stretchr/testify/assert"
)

func TestParseAliasesFromFile(t *testing.T) {
	t.Run("will parse a host file", func(t *testing.T) {
		path := "/usr/share/zoneinfo/tzdata.zi"
		_, err := os.Stat("/usr/share/zoneinfo/tzdata.zi")
		if os.IsNotExist(err) {
			t.Skip("tzdata file does not exist at the expected path, skipping!")
			return
		}

		target := map[string]string{}
		err = zoneinfo.ParseAliasesFromFile(path, target)
		assert.NoError(t, err, "should not have an error parsing the file")
		assert.Contains(t, target, "Asia/Calcutta", "should have a known timezone that changed names")
		assert.Equal(t, "Asia/Kolkata", target["Asia/Calcutta"], "should have the expected new timezone name")
	})
}
