package secrets

type KeyManagement interface {
	Encrypt(input []byte) (keyID, version string, result []byte, _ error)
	Decrypt(keyID, version string, input []byte) (result []byte, _ error)
}
