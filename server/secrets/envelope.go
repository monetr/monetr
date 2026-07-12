package secrets

import (
	"context"
	"log/slog"

	"github.com/benbjohnson/clock"
	"github.com/go-pg/pg/v10"
	"github.com/monetr/monetr/server/models"
)

type Envelope interface {
	// Encrypt will generate a new data encryption key and encrypt the provided
	// plaintext data with that key. It will then wrap the data encryption key
	// using the currently active key encryption key and return the encrypted data
	// structure with all data encrypted.
	Encrypt(ctx context.Context, plaintext []byte) (models.EncryptedData, error)
	// Decrypt will take the encrypted data structure and unwrap the data
	// encryption key using it's key encryption key Id. It will then decrypt the
	// stored data and return it to the caller.
	Decrypt(ctx context.Context, input models.EncryptedData) ([]byte, error)
	// Rewrap takes an already encrypted data structure and unwraps and decrypts
	// it. It will then generate a net new data encryption key and rewrap it
	// against the current active key encryption key and return the data
	// structure.
	Rewrap(ctx context.Context, input models.EncryptedData) (models.EncryptedData, error)
	// ActiveKekId returns the ID of the currently used key encryption key.
	ActiveKekId(ctx context.Context) (models.ID[models.KeyEncryptionKey], error)
	Close() error
}

type envelope struct {
	log   *slog.Logger
	clock clock.Clock
	db    pg.DBI
}
