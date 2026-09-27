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
	// TLS will enable TLS for the connection to the redis server. TLS is also
	// used if any of the certificate paths are provided.
	TLS                bool   `yaml:"tls"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	CACertificatePath  string `yaml:"caCertificatePath"`
	KeyPath            string `yaml:"keyPath"`
	CertificatePath    string `yaml:"certificatePath"`
}
