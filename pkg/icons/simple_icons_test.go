//go:build icons && simple_icons
package icons

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleIconsSearch(t *testing.T) {
	t.Run("amazon", func(t *testing.T) {
		icon, err := SearchIcon("Amazon")
		assert.NoError(t, err, "should not return an error")
		assert.NotNil(t, icon, "should find an amazon icon")
	})
}

