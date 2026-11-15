package encryption

import (
	"context"

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

func newRootKeyUnwrapper(ctx context.Context, cfg config.Encryption) (rootKeyUnwrapper, error) {
	switch models.RootKeyKind(cfg.Kind) {
	case models.RootKeyKindSecretBox:
		secret, err := loadSecretKey(cfg.SecretBox)
		if err != nil {
			return nil, err
		}
		return newSecretboxUnwrapper(secret)
	case models.RootKeyKindAWS:
		return newAWSUnwrapper(ctx, cfg.AWS)
	case models.RootKeyKindOpenBao:
		return newOpenBaoUnwrapper(ctx, cfg.OpenBao)
	case "":
		return nil, errors.New("encryption.kind is not configured")
	default:
		return nil, errors.Errorf("unknown encryption.kind %q, must be secretbox, aws, or openbao", cfg.Kind)
	}
}
