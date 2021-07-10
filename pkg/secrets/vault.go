package secrets

import (
	"context"
	"fmt"
	"github.com/monetr/rest-api/pkg/internal/vault_helper"
	"github.com/sirupsen/logrus"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

var (
	_ PlaidSecretsProvider = &vaultPlaidSecretsProvider{}
)

type vaultPlaidSecretsProvider struct {
	log    *logrus.Entry
	client vault_helper.VaultHelper
}

func NewVaultPlaidSecretsProvider(log *logrus.Entry, client vault_helper.VaultHelper) PlaidSecretsProvider {
	return &vaultPlaidSecretsProvider{
		log:    log,
		client: client,
	}
}

func (v *vaultPlaidSecretsProvider) UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId string, accessToken string) error {
	span := sentry.StartSpan(ctx, "UpdateAccessTokenForPlaidLinkId [VAULT]")
	defer span.Finish()

	path := v.buildPath(accountId, plaidItemId)

	if err := v.client.WriteKV(span.Context(), path, map[string]interface{}{
		"access_token": accessToken,
	}); err != nil {
		return errors.Wrap(err, "failed to store access token in vault")
	}

	return nil
}

func (v *vaultPlaidSecretsProvider) GetAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId string) (accessToken string, _ error) {
	span := sentry.StartSpan(ctx, "GetAccessTokenForPlaidLinkId [VAULT]")
	defer span.Finish()

	result, err := v.client.ReadKV(span.Context(), v.buildPath(accountId, plaidItemId))
	if err != nil {
		return "", errors.Wrap(err, "failed to read access token from vault")
	}

	if result.Data != nil {
		var ok bool
		if accessToken, ok = result.Data["access_token"].(string); ok {
			return accessToken, nil
		}
	}

	return "", errors.Errorf("access token could not be found")
}

func (v *vaultPlaidSecretsProvider) buildPath(accountId uint64, plaidLinkId string) string {
	return fmt.Sprintf("secret/customers/plaid/data/%X/%s", accountId, plaidLinkId)
}

func (v *vaultPlaidSecretsProvider) Close() error {
	return nil
}
