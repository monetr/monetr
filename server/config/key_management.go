package config

// KeyManagement specifies the properties required to securely encrypt and
// decrypt stored secrets. It is not recommended to change providers.
type KeyManagement struct {
	Provider string `yaml:"provider"`
	// AWS provides configuration for using AWS's KMS for encrypting and
	// decrypting secrets.
	AWS AWSKMS `yaml:"aws"`
	// Vault provides configuration for using Vault's Transit API for encrypting
	// and decrypting secrets.
	Vault VaultTransit `yaml:"vault"`
}

type AWSKMS struct {
	AccessKey string  `yaml:"accessKey"`
	Endpoint  *string `yaml:"endpoint"`
	Region    string  `yaml:"region"`
	SecretKey string  `yaml:"secretKey"`

	KeyID string `yaml:"keyID"`
}

type VaultTransit struct {
	// The name of the Vault Transit key to be used for encryption and decryption.
	// This value can be changed but the old key must always be accessible to
	// allow old secrets to be decrypted. But new secrets will be encrypted using
	// the new key. This value cannot be left blank.
	KeyID string `json:"keyID"`
	// AuthMethod tells monetr how to authenticate Vault, potential values are:
	// - `token`: Use a token to authenticate to Vault.
	// - `kubernetes`: Use the container's Kubernetes Service Account Token to
	//                 authenticate to Vault.
	// - `userpass`: Provide a username and password.
	AuthMethod string `yaml:"authMethod"`
	// Token can be specified directly via the config if you are using the token
	// authentication method.
	Token string `yaml:"token"`
	// TokenFile is used when authenticating via Kubernetes, it specifies the path
	// to the Kubernetes service account token file. This file can change over
	// time if needed, actual credentials are issued with short lifetimes and are
	// automatically refreshed using the contents of this file.
	TokenFile string `yaml:"tokenFile"`
	// Vault URL to make requests to.
	Endpoint string `yaml:"endpoint"`
	// Role is required if you are authenticating via `userpass` or `kubernetes`.
	// This role is sent with the authentication request to specify what role the
	// client will have.
	Role string `yaml:"role"`
	// Username is required when you are authenticating via `userpass`.
	Username string `yaml:"username"`
	// Password is required when you are authenticating via `userpass`.
	Password string `yaml:"password"`
	// TLSCertificatePath can be specified if your Vault server is using a custom
	// certificate authority. This path will also be monitored for changes such
	// that certificates can be rotated without downtime.
	TLSCertificatePath string `yaml:"tlsCertificatePath"`
	// Specify the path to the TLS key for Vault.
	TLSKeyPath string `yaml:"tlsKeyPath"`
	// Specify the path to the certificate authority that Vault is using.
	TLSCAPath string `yaml:"tlsCAPath"`
	// Ignores any TLS issues.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
}
