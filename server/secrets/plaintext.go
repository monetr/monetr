package secrets

import "context"

type plaintextKms struct{}

func NewPlaintextKMS() KeyManagement {
	return plaintextKms{}
}

func (plaintextKms) Encrypt(ctx context.Context, input string) (keyId, version *string, result string, _ error) {
	return nil, nil, input, nil
}

func (plaintextKms) Decrypt(ctx context.Context, keyId, version *string, input string) (result string, _ error) {
	return input, nil
}
