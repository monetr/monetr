package encryption

import (
	"context"
	"log/slog"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

// rootKeyWrapper wraps and unwraps root keys for one kind: secretbox, AWS KMS,
// or OpenBao Transit.
type rootKeyWrapper interface {
	Kind() models.RootKeyKind
	// Matches reports whether this wrapper produced the root key wrapper. Same
	// secret key, same KMS key, or same transit key.
	Matches(wrapper models.RootKeyWrapper) bool
	// Wrap encrypts a root key and returns the wrapper to store on its row.
	Wrap(ctx context.Context, rootKey, aad []byte) (models.RootKeyWrapper, error)
	// Unwrap decrypts a root key from a wrapper this wrapper produced.
	Unwrap(ctx context.Context, wrapper models.RootKeyWrapper, aad []byte) ([]byte, error)
}

func newRootKeyWrapper(
	ctx context.Context,
	log *slog.Logger,
	config config.Encryption,
) (rootKeyWrapper, error) {
	switch models.RootKeyKind(config.Kind) {
	case models.RootKeyKindSecretBox:
		secret, err := loadSecretboxKey(log, config.SecretBox)
		if err != nil {
			return nil, err
		}
		return newSecretboxWrapper(secret)
	case models.RootKeyKindAWS:
		return newAWSWrapper(ctx, config.AWS)
	case models.RootKeyKindOpenBao:
		return newOpenBaoWrapper(ctx, config.OpenBao)
	case "":
		return nil, errors.New("encryption.kind is not configured")
	default:
		return nil, errors.Errorf("unknown encryption.kind %q, must be secretbox, aws, or openbao", config.Kind)
	}
}
