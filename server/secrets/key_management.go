package secrets

import (
	"context"
)

//go:generate mockgen -source=key_management.go -package=mockgen -destination=../internal/mockgen/key_management.go KeyManagement
type KeyManagement interface {
	Encrypt(ctx context.Context, input []byte) (keyID, version string, result []byte, _ error)
	Decrypt(ctx context.Context, keyID, version string, input []byte) (result []byte, _ error)
}
