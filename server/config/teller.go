package config

type Teller struct {
	Enabled       bool   `yaml:"enabled"`
	ApplicationId string `yaml:"applicationId"`
	// Certificate is the path to the certificate.pem file provided by teller.io
	// for your client application.
	Certificate string `yaml:"certificate"`
	// PrivateKey is the path to the private_key.pem file provided by teller.io
	// for your client application.
	PrivateKey string `yaml:"privateKey"`
}

func (t Teller) GetEnabled() bool {
	return t.Enabled && t.ApplicationId != ""
}
