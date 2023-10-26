package config_test

import (
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfiguration(t *testing.T) {
	t.Run("login expiration default values", func(t *testing.T) {
		v := viper.GetViper()
		// Don't set a value for login expiration at all.

		configuration := config.LoadConfigurationEx(v)
		assert.NotZero(t, configuration.JWT.LoginExpiration, "Login expiration should not be zero")
		assert.EqualValues(t, 7, configuration.JWT.LoginExpiration, "Login expiration should be 7 days by default")
	})

	t.Run("login expiration non-default values", func(t *testing.T) {
		v := viper.GetViper()
		durationString := "30"
		v.Set("jwt.loginExpiration", durationString)

		configuration := config.LoadConfigurationEx(v)
		assert.NotZero(t, configuration.JWT.LoginExpiration, "Login expiration should not be zero")
		assert.EqualValues(t, 30, configuration.JWT.LoginExpiration, "when specified, the login expiration should not be the default")
	})
}
