package config

type Links struct {
	// MaxNumberOfLinks defines the max number of active data connector links a
	// single account can have. If this is set to 0 then there is no limit.
	MaxNumberOfLinks int `yaml:"maxNumberOfLinks"`
}
