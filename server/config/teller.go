package config

type Teller struct {
	Enabled       bool   `yaml:"enabled"`
	Environment   string `yaml:"environment"`
	ApplicationId string `yaml:"applicationId"`
	// Certificate is the path to the certificate.pem file provided by teller.io
	// for your client application.
	Certificate string `yaml:"certificate"`
	// PrivateKey is the path to the private_key.pem file provided by teller.io
	// for your client application.
	PrivateKey      string `yaml:"privateKey"`
	TokenSigningKey string `yaml:"tokenSigningKey"`
	// When webhooks are received the will have a Teller-Signature header, this
	// signature header is verified using one of the secrets in this array. If
	// none of the secrets match then the webhook is rejected.
	WebhookSigningSecret []string `yaml:"webhookSigningSecret"`
}

func (t Teller) GetEnabled() bool {
	return t.Enabled && t.ApplicationId != ""
}
