package secrets

import "context"

type plaintextKms struct{}

func NewPlaintextKMS() KeyManagement {
	return plaintextKms{}
}

func (plaintextKms) Encrypt(ctx context.Context, input []byte) (keyId, version *string, result []byte, _ error) {
	return nil, nil, input, nil
}

func (plaintextKms) Decrypt(ctx context.Context, keyId, version *string, input []byte) (result []byte, _ error) {
	return input, nil
}
