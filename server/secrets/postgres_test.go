package secrets_test

import (
	"context"
	"encoding/base64"
	"math"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang/mock/gomock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
)

func TestPostgresSecretStorage_Store(t *testing.T) {
	t.Run("account does not exist", func(t *testing.T) {
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		accessToken := gofakeit.UUID()

		provider := secrets.NewPostgresSecretsStorage(
			log,
			db,
			secrets.NewPlaintextKMS(),
		)
		err := provider.Store(ctx, &secrets.Data{
			AccountId: math.MaxInt64,
			Kind:      models.PlaidSecretKind,
			Secret:    accessToken,
		})
		assert.EqualError(t, err, `failed to update access token: ERROR #23503 insert or update on table "plaid_tokens" violates foreign key constraint "fk_plaid_tokens_account"`)
	})

	t.Run("first write", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		accessToken := gofakeit.UUID()

		provider := secrets.NewPostgresSecretsStorage(
			log,
			db,
			secrets.NewPlaintextKMS(),
		)
		secret := secrets.Data{
			AccountId: user.AccountId,
			Kind:      models.PlaidSecretKind,
			Secret:    accessToken,
		}
		err := provider.Store(ctx, &secret)
		assert.NoError(t, err, "must be able to write access token for the first time")

		data, err := provider.Read(ctx, secret.SecretId, secret.AccountId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessToken, data.Secret, "retrieved token must match the one written")
	})

	t.Run("first write kms", func(t *testing.T) {
		clock := clock.NewMock()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
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
				&keyName,  // Key name
				&version,  // Key version
				encrypted, // Encrypted value
				nil,       // Error
			).
			MaxTimes(1)

		provider := secrets.NewPostgresSecretsStorage(log, db, kms)
		secret := secrets.Data{
			AccountId: user.AccountId,
			Kind:      models.PlaidSecretKind,
			Secret:    accessToken,
		}
		err := provider.Store(ctx, &secret)
		assert.NoError(t, err, "must be able to write access token for the first time")

		kms.EXPECT().
			Decrypt(
				gomock.Any(),
				gomock.Eq(keyName),
				gomock.Eq(version),
				gomock.Eq(encrypted),
			).
			Return(
				[]byte(accessToken), // Decrypted access token
				nil,                 // Error
			).
			MaxTimes(1)

		result, err := provider.Read(ctx, secret.SecretId, secret.AccountId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessToken, result.Secret, "retrieved token must match the one written")
	})

	t.Run("handle changes", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()

		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		accessToken := gofakeit.UUID()

		provider := secrets.NewPostgresSecretsStorage(
			log,
			db,
			secrets.NewPlaintextKMS(),
		)
		secret := secrets.Data{
			AccountId: user.AccountId,
			Kind:      models.PlaidSecretKind,
			Secret:    accessToken,
		}
		err := provider.Store(ctx, &secret)
		assert.NoError(t, err, "must be able to write access token for the first time")

		result, err := provider.Read(ctx, secret.SecretId, secret.AccountId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessToken, result.Secret, "retrieved token must match the one written")

		accessTokenUpdated := gofakeit.UUID()
		assert.NotEqual(t, accessToken, accessTokenUpdated, "make sure the new token does not match the old one")

		secret.Secret = accessTokenUpdated
		err = provider.Store(ctx, &secret)
		assert.NoError(t, err, "must be able to update an existing access token")

		result, err = provider.Read(ctx, secret.SecretId, secret.AccountId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessTokenUpdated, result.Secret, "retrieved token must match the second one written")
	})
}
