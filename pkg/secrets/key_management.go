package secrets

import (
	"context"
)

type KeyManagement interface {
	Encrypt(ctx context.Context, input []byte) (keyID, version string, result []byte, _ error)
	Decrypt(ctx context.Context, keyID, version string, input []byte) (result []byte, _ error)
}
