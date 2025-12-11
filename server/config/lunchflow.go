package config

type LunchFlow struct {
	// Enabled just determines whether or not Lunch Flow will be an option to
	// configure in the UI. This defaults to true as it requires no additional
	// configuration here for self-hosted users.
	Enabled bool `yaml:"enabled"`
}
