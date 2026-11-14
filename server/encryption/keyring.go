package encryption

import (
	"context"
	"crypto/rand"
	"sync"

	"github.com/monetr/monetr/server/models"
	"golang.org/x/crypto/chacha20poly1305"
)

var (
	_ models.DataKeyWrapper = &Keyring{}
)

type Keyring struct {
	lock      sync.RWMutex
	activeKEK models.ID[models.KeyEncryptionKey]
	keks      map[models.ID[models.KeyEncryptionKey]][]byte
}

// UnwrapDataKey implements [models.DataKeyWrapper].
func (k *Keyring) UnwrapDataKey(ctx context.Context, kekId models.ID[models.KeyEncryptionKey], wrapped []byte) ([]byte, error) {
	panic("unimplemented")
}

// WrapDataKey implements [models.DataKeyWrapper].
func (k *Keyring) WrapDataKey(
	ctx context.Context,
	dek []byte,
) (models.ID[models.KeyEncryptionKey], []byte, error) {
	k.lock.RLock()
	kekId := k.activeKEK
	kek := k.keks[kekId]
	k.lock.RUnlock()

	aead, err := chacha20poly1305.NewX(kek)
	if err != nil {
		return "", nil, err
	}
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	rand.Read(nonce)
	// AAD is the KEK ID, so a wrapped DEK can't be relabeled as belonging to a
	// different KEK. Result is nonce || ciphertext.
	return kekId, aead.Seal(nonce, nonce, dek, []byte(kekId)), nil
}
