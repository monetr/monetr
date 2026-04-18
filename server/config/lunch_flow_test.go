package config_test

import (
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/stretchr/testify/assert"
)

func TestLunchFlow_ValidateConfig(t *testing.T) {
	t.Run("disabled is a no-op", func(t *testing.T) {
		configuration := config.LunchFlow{
			Enabled:        false,
			AllowedApiUrls: []string{"not a valid url"},
		}
		assert.NoError(t, configuration.ValidateConfig())
	})

	t.Run("no allowed urls is a no-op", func(t *testing.T) {
		configuration := config.LunchFlow{
			Enabled: true,
		}
		assert.NoError(t, configuration.ValidateConfig())
	})

	t.Run("valid urls are accepted", func(t *testing.T) {
		configuration := config.LunchFlow{
			Enabled: true,
			AllowedApiUrls: []string{
				"https://lunchflow.app/api/v1",
				"http://lunchflow.app/api/v1",
			},
		}
		assert.NoError(t, configuration.ValidateConfig())
	})

	t.Run("unparseable url is rejected", func(t *testing.T) {
		configuration := config.LunchFlow{
			Enabled:        true,
			AllowedApiUrls: []string{"https://lunchflow.app/%zz"},
		}
		assert.EqualError(t, configuration.ValidateConfig(), `configured Lunch Flow url (https://lunchflow.app/%zz) is not valid: parse "https://lunchflow.app/%zz": invalid URL escape "%zz"`)
	})

	t.Run("url with query parameters is rejected", func(t *testing.T) {
		configuration := config.LunchFlow{
			Enabled:        true,
			AllowedApiUrls: []string{"https://lunchflow.app/api/v1?token=secret"},
		}
		assert.EqualError(t, configuration.ValidateConfig(), "Lunch Flow url (https://lunchflow.app/api/v1?token=secret) cannot contain query parameters")
	})

	t.Run("invalid url in config", func(t *testing.T) {
		configuration := config.LunchFlow{
			Enabled:        true,
			AllowedApiUrls: []string{"example.com"},
		}
		assert.EqualError(t, configuration.ValidateConfig(), "Lunch Flow url (example.com) must use an http or https scheme")
	})
}

func TestLunchFlow_IsAllowedApiUrl(t *testing.T) {
	t.Run("exact match is allowed", func(t *testing.T) {
		configuration := config.LunchFlow{
			AllowedApiUrls: []string{"https://lunchflow.app/api/v1"},
		}
		assert.True(t, configuration.IsAllowedApiUrl("https://lunchflow.app/api/v1"))
	})

	t.Run("mismatch is rejected", func(t *testing.T) {
		configuration := config.LunchFlow{
			AllowedApiUrls: []string{"https://lunchflow.app/api/v1"},
		}
		assert.False(t, configuration.IsAllowedApiUrl("http://169.254.169.254/latest/meta-data"))
		assert.False(t, configuration.IsAllowedApiUrl("http://127.0.0.1"))
		assert.False(t, configuration.IsAllowedApiUrl("https://lunchflow.app/api/v2"))
	})

	t.Run("empty list rejects everything", func(t *testing.T) {
		configuration := config.LunchFlow{}
		assert.False(t, configuration.IsAllowedApiUrl("https://lunchflow.app/api/v1"))
		assert.False(t, configuration.IsAllowedApiUrl(""))
	})

	t.Run("multiple entries match any of them", func(t *testing.T) {
		configuration := config.LunchFlow{
			AllowedApiUrls: []string{
				"https://lunchflow.app/api/v1",
				"https://lunchflow.compatible.app/api/v1",
			},
		}
		assert.True(t, configuration.IsAllowedApiUrl("https://lunchflow.app/api/v1"))
		assert.True(t, configuration.IsAllowedApiUrl("https://lunchflow.compatible.app/api/v1"))
		assert.False(t, configuration.IsAllowedApiUrl("https://other.lunchflow.app/api/v1"))
	})
}
