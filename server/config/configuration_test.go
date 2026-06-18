package config_test

import (
	"strings"
	"testing"

	"github.com/monetr/monetr/server/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigurationEx_BlockedDomains(t *testing.T) {
	// These tests exist because BlockedDomains is a myownsanity.Set, which is a
	// map under the hood. Viper uses mapstructure (NOT encoding/json) to load the
	// config, so the Set's json unmarshaler never gets called and without a
	// custom decode hook the whole config load panics trying to shove a list into
	// a map. See stringSliceToSetHookFunc.
	t.Run("loads a yaml list into the set", func(t *testing.T) {
		yamlData := `
email:
  enabled: true
  blockedDomains:
    - mailinator.com
    - guerrillamail.com
`
		v := viper.New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader(yamlData)), "must be able to read the yaml")

		conf := config.LoadConfigurationEx(v)
		assert.Len(t, conf.Email.BlockedDomains, 2, "both domains from the yaml list should be loaded")
		assert.True(t, conf.Email.BlockedDomains.Has("mailinator.com"), "should contain the first domain")
		assert.True(t, conf.Email.BlockedDomains.Has("guerrillamail.com"), "should contain the second domain")
	})

	t.Run("loads a comma separated env var into the set", func(t *testing.T) {
		// The MONETR_EMAIL_BLOCKED_DOMAINS env var comes through as a single comma
		// separated string rather than a list, so the hook has to split it.
		t.Setenv("MONETR_EMAIL_BLOCKED_DOMAINS", "mailinator.com, guerrillamail.com")

		v := viper.New()
		v.SetConfigType("yaml")
		v.MustBindEnv("Email.BlockedDomains", "MONETR_EMAIL_BLOCKED_DOMAINS")
		require.NoError(t, v.ReadConfig(strings.NewReader("")), "must be able to read empty config")

		conf := config.LoadConfigurationEx(v)
		assert.Len(t, conf.Email.BlockedDomains, 2, "both domains from the env var should be loaded")
		assert.True(t, conf.Email.BlockedDomains.Has("mailinator.com"), "should contain the first domain")
		assert.True(t, conf.Email.BlockedDomains.Has("guerrillamail.com"), "and should trim the whitespace after the comma")
	})

	t.Run("an absent blocklist loads as an empty set", func(t *testing.T) {
		v := viper.New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader("email:\n  enabled: true\n")), "must be able to read the yaml")

		conf := config.LoadConfigurationEx(v)
		assert.Empty(t, conf.Email.BlockedDomains, "a config without a blocklist should load an empty set, not panic")
	})
}

func TestLoadConfigurationEx(t *testing.T) {
	// When we added the Set decode hook we had to pass viper.DecodeHook, which
	// completely replaces viper's defaults. These tests are just a sanity check
	// that the rest of the config still loads like it always did, in particular
	// that we did not drop the default duration and slice hooks on the floor when
	// we spelled them out by hand. This is NOT trying to cover the whole config,
	// just one of each interesting kind of value.
	t.Run("a bool loads", func(t *testing.T) {
		v := viper.New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader("allowSignUp: true\n")), "must be able to read the yaml")

		conf := config.LoadConfigurationEx(v)
		assert.True(t, conf.AllowSignUp, "the bool should load from the yaml")
	})

	t.Run("an integer loads", func(t *testing.T) {
		yamlData := `
postgreSql:
  port: 6000
`
		v := viper.New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader(yamlData)), "must be able to read the yaml")

		conf := config.LoadConfigurationEx(v)
		assert.Equal(t, 6000, conf.PostgreSQL.Port, "the integer should load from the yaml")
	})

	t.Run("a string loads", func(t *testing.T) {
		v := viper.New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader("environment: production\n")), "must be able to read the yaml")

		conf := config.LoadConfigurationEx(v)
		assert.Equal(t, "production", conf.Environment, "the string should load from the yaml")
	})

	t.Run("a float loads", func(t *testing.T) {
		yamlData := `
sentry:
  sampleRate: 0.25
`
		v := viper.New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader(yamlData)), "must be able to read the yaml")

		conf := config.LoadConfigurationEx(v)
		assert.Equal(t, 0.25, conf.Sentry.SampleRate, "the float should load from the yaml")
	})

	t.Run("a duration loads", func(t *testing.T) {
		// Durations come in through viper's StringToTimeDurationHookFunc, one of
		// the defaults we had to re-add by hand, so this guards against us having
		// dropped it.
		yamlData := `
proofOfWork:
  lifetime: 30s
`
		v := viper.New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader(yamlData)), "must be able to read the yaml")

		conf := config.LoadConfigurationEx(v)
		assert.Equal(t, "30s", conf.ProofOfWork.Lifetime.String(), "the duration should still parse")
	})

	t.Run("a string slice loads", func(t *testing.T) {
		// Slices come in through StringToSliceHookFunc, the other default we had
		// to re-add, so this is the matching regression guard for it.
		yamlData := `
cors:
  allowedOrigins:
    - https://app.example.com
    - https://example.com
`
		v := viper.New()
		v.SetConfigType("yaml")
		require.NoError(t, v.ReadConfig(strings.NewReader(yamlData)), "must be able to read the yaml")

		conf := config.LoadConfigurationEx(v)
		assert.ElementsMatch(t, []string{
			"https://app.example.com",
			"https://example.com",
		}, conf.CORS.AllowedOrigins, "the string slice should load from the yaml")
	})
}
