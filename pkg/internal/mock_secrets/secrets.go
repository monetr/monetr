package mock_secrets

import (
	"context"
	"fmt"
	"github.com/monetrapp/rest-api/pkg/secrets"
	"github.com/pkg/errors"
)

var (
	_ secrets.PlaidSecretsProvider = &MockPlaidSecrets{}
)

type MockPlaidSecrets struct {
	store map[string]string
}

func NewMockPlaidSecrets() *MockPlaidSecrets {
	return &MockPlaidSecrets{
		store: map[string]string{},
	}
}

func (m *MockPlaidSecrets) WithSecret(accountId uint64, itemId, accessToken string) *MockPlaidSecrets {
	m.store[m.getKey(accountId, itemId)] = accessToken

	return m
}

func (m *MockPlaidSecrets) getKey(accountId uint64, itemId string) string {
	return fmt.Sprintf("secret/plaid/clients/data/%X/%s", accountId, itemId)
}

func (m *MockPlaidSecrets) UpdateAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId, accessToken string) error {
	m.store[m.getKey(accountId, plaidItemId)] = accessToken

	return nil
}

func (m *MockPlaidSecrets) GetAccessTokenForPlaidLinkId(ctx context.Context, accountId uint64, plaidItemId string) (accessToken string, err error) {
	accessToken, ok := m.store[m.getKey(accountId, plaidItemId)]
	if !ok {
		return "", errors.Errorf("failed to retrieve access token")
	}

	return accessToken, nil
}

func (m *MockPlaidSecrets) Close() error {
	return nil
}
