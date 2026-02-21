package config

// Redis defines the config used to connect to a redis for our worker pool. If
// these are left blank or default then we will instead use a mock redis pool
// that is internal only. This is fine for single instance deployments, but
// anytime more than one instance of the API is running a redis instance will be
// required.
type Redis struct {
	Enabled  bool   `yaml:"enabled"`
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	Database int    `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
