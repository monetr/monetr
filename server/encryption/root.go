package encryption

import (
	"context"
	"log/slog"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

// rootKeyUnwrapper opens the root key wrappers for one configured kind.
type rootKeyUnwrapper interface {
	Kind() models.RootKeyKind
	// Matches reports whether this unwrapper produced the wrapper. Same
	// secret key, same KMS key, or same transit key.
	Matches(wrapper models.RootKeyWrapper) bool
	Unwrap(ctx context.Context, wrapper models.RootKeyWrapper, aad []byte) ([]byte, error)
}

func newRootKeyUnwrapper(
	ctx context.Context,
	log *slog.Logger,
	config config.Encryption,
) (rootKeyUnwrapper, error) {
	switch models.RootKeyKind(config.Kind) {
	case models.RootKeyKindSecretBox:
		secret, err := loadSecretboxKey(log, config.SecretBox)
		if err != nil {
			return nil, err
		}
		return newSecretboxUnwrapper(secret)
	case models.RootKeyKindAWS:
		return newAWSUnwrapper(ctx, config.AWS)
	case models.RootKeyKindOpenBao:
		return newOpenBaoUnwrapper(ctx, config.OpenBao)
	case "":
		return nil, errors.New("encryption.kind is not configured")
	default:
		return nil, errors.Errorf("unknown encryption.kind %q, must be secretbox, aws, or openbao", config.Kind)
	}
}
