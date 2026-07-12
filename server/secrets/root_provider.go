package secrets

import (
	"context"

	"github.com/monetr/monetr/server/models"
)

type RootKeyProvider interface {
	RootKeyID() string

	WrapKey(ctx context.Context, kekId models.ID[models.KeyEncryptionKey], kek []byte) (wrapped, integity []byte, err error)
	UnwrapKey(ctx context.Context, kekId models.ID[models.KeyEncryptionKey], rootKeyId string, wrapped, integrity []byte) (kek []byte, err error)
	Close() error
}
