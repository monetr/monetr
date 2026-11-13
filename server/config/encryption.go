package config

// Encryption configures monetr's secure storage, it is used to encrypt secrets
// or other sensitive pieces of information at rest. It replaces monetr's
// existing secrets implementation.
type Encryption struct {
	// Kind is one of `secretbox` `aws` or `openbao`. Secretbox requires that the
	// secret key is available to the monetr process on startup and decrypts the
	// root key that monetr uses.
	Kind string `yaml:"kind"`
	// SecretBox configures the secret or secret file loading at startup and is
	// suitable for self hosted environments.
	SecretBox SecretBox `yaml:"secretBox"`
	// AWS configures AWS KMS as the thing that will decrypt the initial keys
	// monetr uses for encryption. This allows you to keep the secret key that
	// monetr uses out of the process or filesystem.
	AWS AWSKMS `yaml:"aws"`
	// OpenBao configures monetr to start and decrypt the initial keys using
	// OpenBao Transit instead of requiring the secret key be provided to monetr
	// to start.
	OpenBao OpenBaoTransit `yaml:"openbao"`
}

type SecretBox struct {
	Secret     *string `yaml:"secret"`
	SecretFile *string `yaml:"secretFile"`
}
