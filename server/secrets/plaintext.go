package secrets

import "context"

type plaintextKms struct{}

func NewPlaintextKMS() KeyManagement {
	return plaintextKms{}
}

func (plaintextKms) Encrypt(_ context.Context, input string) (keyId, version *string, result string, _ error) {
	return nil, nil, input, nil
}

func (plaintextKms) Decrypt(_ context.Context, _, _ *string, input string) (result string, _ error) {
	return input, nil
}
