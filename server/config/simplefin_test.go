package config_test

import (
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/stretchr/testify/assert"
)

func TestSimpleFin_ValidateConfig(t *testing.T) {
	t.Run("disabled is a no-op", func(t *testing.T) {
		configuration := config.SimpleFin{
			Enabled: false,
			AllowedHosts: []string{
				"https://not a host",
			},
		}
		assert.NoError(t, configuration.ValidateConfig())
	})

	t.Run("no allowed hosts is a no-op", func(t *testing.T) {
		configuration := config.SimpleFin{
			Enabled: true,
		}
		assert.NoError(t, configuration.ValidateConfig())
	})

	t.Run("valid hosts are accepted", func(t *testing.T) {
		configuration := config.SimpleFin{
			Enabled: true,
			AllowedHosts: []string{
				"bridge.simplefin.org",
				"beta-bridge.simplefin.org",
				"simplefin.example.com:8443",
			},
		}
		assert.NoError(t, configuration.ValidateConfig())
	})

	t.Run("host with a scheme is rejected", func(t *testing.T) {
		configuration := config.SimpleFin{
			Enabled: true,
			AllowedHosts: []string{
				"https://bridge.simplefin.org",
			},
		}
		assert.EqualError(t, configuration.ValidateConfig(), "SimpleFIN host (https://bridge.simplefin.org) must only be a host and an optional port")
	})

	t.Run("host with a path is rejected", func(t *testing.T) {
		configuration := config.SimpleFin{
			Enabled: true,
			AllowedHosts: []string{
				"bridge.simplefin.org/simplefin",
			},
		}
		assert.EqualError(t, configuration.ValidateConfig(), "SimpleFIN host (bridge.simplefin.org/simplefin) must only be a host and an optional port")
	})

	t.Run("empty host is rejected", func(t *testing.T) {
		configuration := config.SimpleFin{
			Enabled: true,
			AllowedHosts: []string{
				"",
			},
		}
		assert.EqualError(t, configuration.ValidateConfig(), "SimpleFIN host () must only be a host and an optional port")
	})

	t.Run("invalid port is rejected", func(t *testing.T) {
		configuration := config.SimpleFin{
			Enabled: true,
			AllowedHosts: []string{
				"bridge.simplefin.org:abc",
			},
		}
		assert.EqualError(t, configuration.ValidateConfig(), `configured SimpleFIN host (bridge.simplefin.org:abc) is not valid: parse "https://bridge.simplefin.org:abc": invalid port ":abc" after host`)
	})
}

func TestSimpleFin_IsAllowedUrl(t *testing.T) {
	t.Run("allowed host", func(t *testing.T) {
		configuration := config.SimpleFin{
			AllowedHosts: []string{
				"bridge.simplefin.org",
			},
		}
		assert.True(t, configuration.IsAllowedUrl("https://bridge.simplefin.org/simplefin/claim/abc123"), "claim url should be allowed")
		assert.True(t, configuration.IsAllowedUrl("https://user:pass@bridge.simplefin.org/simplefin"), "access url should be allowed")
	})

	t.Run("wrong host is rejected", func(t *testing.T) {
		configuration := config.SimpleFin{
			AllowedHosts: []string{
				"bridge.simplefin.org",
			},
		}
		assert.False(t, configuration.IsAllowedUrl("https://169.254.169.254/latest/meta-data"))
		assert.False(t, configuration.IsAllowedUrl("https://127.0.0.1/simplefin"))
		assert.False(t, configuration.IsAllowedUrl("https://bridge.simplefin.org.example.com/simplefin"))
		assert.False(t, configuration.IsAllowedUrl("https://beta-bridge.simplefin.org/simplefin"))
	})

	t.Run("http is rejected", func(t *testing.T) {
		configuration := config.SimpleFin{
			AllowedHosts: []string{
				"bridge.simplefin.org",
			},
		}
		assert.False(t, configuration.IsAllowedUrl("http://bridge.simplefin.org/simplefin"))
	})

	t.Run("port mismatch is rejected", func(t *testing.T) {
		configuration := config.SimpleFin{
			AllowedHosts: []string{
				"bridge.simplefin.org",
			},
		}
		assert.False(t, configuration.IsAllowedUrl("https://bridge.simplefin.org:8443/simplefin"))
	})

	t.Run("host is case insensitive", func(t *testing.T) {
		configuration := config.SimpleFin{
			AllowedHosts: []string{
				"bridge.simplefin.org",
			},
		}
		assert.True(t, configuration.IsAllowedUrl("HTTPS://Bridge.SimpleFIN.org/simplefin"))
	})

	t.Run("empty list rejects everything", func(t *testing.T) {
		configuration := config.SimpleFin{}
		assert.False(t, configuration.IsAllowedUrl("https://bridge.simplefin.org/simplefin"))
	})
}
