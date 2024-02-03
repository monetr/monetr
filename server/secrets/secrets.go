package secrets

import (
	"context"

	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("access token not found")
)

type Data struct {
	SecretId  uint64            `json:"-"`
	AccountId uint64            `json:"-"`
	Kind      models.SecretKind `json:"-"`
	Secret    string            `json:"-"`
}

type SecretsProvider interface {
	Store(ctx context.Context, secret *Data) error
	Read(ctx context.Context, secretId, accountId uint64) (*Data, error)
	Delete(ctx context.Context, secretId, accountId uint64) error
}
