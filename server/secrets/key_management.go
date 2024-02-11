package secrets

import (
	"context"
)

//go:generate mockgen -source=key_management.go -package=mockgen -destination=../internal/mockgen/key_management.go KeyManagement
type KeyManagement interface {
	Encrypt(ctx context.Context, input string) (keyID, version *string, result string, _ error)
	Decrypt(ctx context.Context, keyID, version *string, input string) (result string, _ error)
}
