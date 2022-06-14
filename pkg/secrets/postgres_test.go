package secrets

import (
	"context"
	"math"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/pkg/internal/fixtures"
	"github.com/monetr/monetr/pkg/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestPostgresPlaidSecretProvider_UpdateAccessTokenForPlaidLinkId(t *testing.T) {
	t.Run("account does not exist", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		plaidItemId := gofakeit.UUID()
		accessToken := gofakeit.UUID()

		provider := NewPostgresPlaidSecretsProvider(log, db, nil)
		err := provider.UpdateAccessTokenForPlaidLinkId(ctx, math.MaxInt64, plaidItemId, accessToken)
		assert.EqualError(t, err, `failed to update access token: ERROR #23503 insert or update on table "plaid_tokens" violates foreign key constraint "fk_plaid_tokens_account"`)
	})

	t.Run("first write", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidItemId := gofakeit.UUID()
		accessToken := gofakeit.UUID()

		provider := NewPostgresPlaidSecretsProvider(log, db, nil)
		err := provider.UpdateAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId, accessToken)
		assert.NoError(t, err, "must be able to write access token for the first time")

		token, err := provider.GetAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessToken, token, "retrieved token must match the one written")
	})

	t.Run("handle changes", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidItemId := gofakeit.UUID()
		accessToken := gofakeit.UUID()

		provider := NewPostgresPlaidSecretsProvider(log, db, nil)
		err := provider.UpdateAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId, accessToken)
		assert.NoError(t, err, "must be able to write access token for the first time")

		token, err := provider.GetAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessToken, token, "retrieved token must match the one written")

		accessTokenUpdated := gofakeit.UUID()
		assert.NotEqual(t, accessToken, accessTokenUpdated, "make sure the new token does not match the old one")

		err = provider.UpdateAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId, accessTokenUpdated)
		assert.NoError(t, err, "must be able to update an existing access token")

		token, err = provider.GetAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessTokenUpdated, token, "retrieved token must match the second one written")
	})
}

func TestPostgresPlaidSecretProvider_RemoveAccessTokenForPlaidLink(t *testing.T) {
}
