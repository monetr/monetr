package secrets_test

import (
	"context"
	"encoding/base64"
	"math"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
)

func TestPostgresPlaidSecretProvider_UpdateAccessTokenForPlaidLinkId(t *testing.T) {
	t.Run("account does not exist", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		plaidItemId := gofakeit.UUID()
		accessToken := gofakeit.UUID()

		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)
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

		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)
		err := provider.UpdateAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId, accessToken)
		assert.NoError(t, err, "must be able to write access token for the first time")

		token, err := provider.GetAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessToken, token, "retrieved token must match the one written")
	})

	t.Run("first write kms", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		user, _ := fixtures.GivenIHaveABasicAccount(t)
		plaidItemId := gofakeit.UUID()
		accessToken := gofakeit.UUID()

		kms := mockgen.NewMockKeyManagement(ctrl)

		encrypted := []byte(base64.StdEncoding.EncodeToString([]byte(accessToken)))
		version := "1"
		keyName := "project/us-east1/key"
		kms.EXPECT().
			Encrypt(
				gomock.Any(),
				gomock.Eq([]byte(accessToken)),
			).
			Return(
				version,   // Key version
				keyName,   // Key name
				encrypted, // Encrypted value
				nil,       // Error
			).
			MaxTimes(1)

		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, kms)
		err := provider.UpdateAccessTokenForPlaidLinkId(ctx, user.AccountId, plaidItemId, accessToken)
		assert.NoError(t, err, "must be able to write access token for the first time")

		kms.EXPECT().
			Decrypt(
				gomock.Any(),
				gomock.Eq(version),
				gomock.Eq(keyName),
				gomock.Eq(encrypted),
			).
			Return(
				[]byte(accessToken), // Decrypted access token
				nil,                 // Error
			).
			MaxTimes(1)

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

		provider := secrets.NewPostgresPlaidSecretsProvider(log, db, nil)
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
