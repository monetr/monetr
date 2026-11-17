package encryption

import (
	"context"
	"log/slog"
	"time"

	"github.com/monetr/monetr/server/config"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/pkg/errors"
)

type openBaoWrapper struct {
	client *secrets.OpenBaoTransit // reuse the auth and refresh setup from secrets.NewOpenBaoTransit
	keyId  string
}

func newOpenBaoWrapper(
	ctx context.Context,
	log *slog.Logger,
	config config.OpenBaoTransit,
) (*openBaoWrapper, error) {
	secrets.NewOpenBaoTransit(ctx, secrets.OpenBaoTransitConfig{
		Log:                log,
		KeyID:              config.KeyID,
		Address:            config.Endpoint,
		Role:               config.Role,
		AuthMethod:         config.AuthMethod,
		Token:              config.Token,
		TokenFile:          config.TokenFile,
		Username:           config.Username,
		Password:           config.Password,
		Timeout:            15 * time.Second,
		TLSCertificatePath: config.TLSCertificatePath,
		TLSKeyPath:         config.TLSKeyPath,
		TLSCAPath:          config.TLSCAPath,
		InsecureSkipVerify: config.InsecureSkipVerify,
		IdleConnTimeout:    15 * time.Second,
	})
}

func (o *openBaoWrapper) Kind() models.RootKeyKind {
	return models.RootKeyKindOpenBao
}

func (o *openBaoWrapper) Matches(w models.RootKeyWrapper) bool {
	return w.Kind == models.RootKeyKindOpenBao && w.Reference == o.keyId
}

func (o *openBaoWrapper) Wrap(
	ctx context.Context,
	rootKey, aad []byte,
) (models.RootKeyWrapper, error) {
	return models.RootKeyWrapper{}, errors.New("wrapping root keys with OpenBao is not implemented yet")
}

func (o *openBaoWrapper) Unwrap(ctx context.Context, w models.RootKeyWrapper, aad []byte) ([]byte, error) {
	panic("not implemented")
}
