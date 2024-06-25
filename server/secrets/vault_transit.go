package secrets

import (
	"context"

	"github.com/sirupsen/logrus"
)

var (
	_ KeyManagement = &VaultTransit{}
)

type VaultTransitConfig struct {
	log *logrus.Entry
}

type VaultTransit struct {
	log *logrus.Entry
}

func NewVaultTransit(ctx context.Context, config VaultTransitConfig) (KeyManagement, error) {
	return &VaultTransit{
		log: config.log,
	}, nil
}

// Decrypt implements KeyManagement.
func (*VaultTransit) Decrypt(
	ctx context.Context,
	keyID *string,
	version *string,
	input string,
) (result string, _ error) {
	panic("unimplemented")
}

// Encrypt implements KeyManagement.
func (*VaultTransit) Encrypt(
	ctx context.Context,
	input string,
) (keyID *string, version *string, result string, _ error) {
	panic("unimplemented")
}
