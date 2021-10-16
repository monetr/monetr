package secrets

import (
	"context"
	"fmt"

	"github.com/monetr/monetr/pkg/internal/vault_helper"
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

		if data, ok := result.Data["data"].(map[string]interface{}); ok {
			if accessToken, ok = data["access_token"].(string); ok {
				return accessToken, nil
			}
		}
	}

	return "", errors.Errorf("access token could not be found")
}

func (v *vaultPlaidSecretsProvider) RemoveAccessTokenForPlaidLink(ctx context.Context, accountId uint64, plaidItemId string) error {
	span := sentry.StartSpan(ctx, "RemoveAccessTokenForPlaidLink [VAULT]")
	defer span.Finish()
	span.Data = map[string]interface{}{
		"itemId": plaidItemId,
	}

	if err := v.client.DeleteKV(span.Context(), v.buildPath(accountId, plaidItemId)); err != nil {
		return errors.Wrap(err, "failed to remove Plaid access token")
	}

	return nil
}

func (v *vaultPlaidSecretsProvider) buildPath(accountId uint64, plaidItemId string) string {
	return fmt.Sprintf("secret/customers/plaid/data/%X/%s", accountId, plaidItemId)
}

func (v *vaultPlaidSecretsProvider) Close() error {
	return nil
}
