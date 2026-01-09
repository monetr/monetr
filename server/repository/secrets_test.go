package repository_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/mockgen"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/monetr/monetr/server/secrets"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSecretsRepository_Store(t *testing.T) {
	t.Run("account does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := secrets.NewPlaintextKMS()
		ctx := context.Background()

		repo := repository.NewSecretsRepository(log, clock, db, kms, "acct_bogus")

		err := repo.Store(ctx, &repository.SecretData{
			Kind:  models.SecretKindPlaid,
			Value: gofakeit.UUID(),
		})
		assert.EqualError(t, err, `failed to store secret: ERROR #23503 insert or update on table "secrets" violates foreign key constraint "fk_secrets_account"`)
	})

	t.Run("first write", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := secrets.NewPlaintextKMS()
		ctx := context.Background()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		accessToken := gofakeit.UUID()

		repo := repository.NewSecretsRepository(log, clock, db, kms, user.AccountId)

		secret := repository.SecretData{
			Kind:  models.SecretKindPlaid,
			Value: accessToken,
		}
		err := repo.Store(ctx, &secret)
		assert.NoError(t, err, "should be able to store the secret successfully")
		assert.NotZero(t, secret.SecretId, "secret Id should now be set")

		result, err := repo.Read(ctx, secret.SecretId)
		assert.NoError(t, err, "should be able to store the secret successfully")
		assert.Equal(t, secret.Value, result.Value, "should read the same value back")
	})

	t.Run("first write with kms", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		ctx := context.Background()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		accessToken := gofakeit.UUID()

		kms := mockgen.NewMockKeyManagement(ctrl)

		encrypted := base64.StdEncoding.EncodeToString([]byte(accessToken))
		version := "1"
		keyName := "project/us-east1/key"
		kms.EXPECT().
			Encrypt(
				gomock.Any(),
				gomock.Eq(accessToken),
			).
			Return(
				&keyName,  // Key name
				&version,  // Key version
				encrypted, // Encrypted value
				nil,       // Error
			).
			MaxTimes(1)

		repo := repository.NewSecretsRepository(log, clock, db, kms, user.AccountId)

		secret := repository.SecretData{
			Kind:  models.SecretKindPlaid,
			Value: accessToken,
		}
		err := repo.Store(ctx, &secret)
		assert.NoError(t, err, "should be able to store the secret successfully")
		assert.NotZero(t, secret.SecretId, "secret Id should now be set")

		kms.EXPECT().
			Decrypt(
				gomock.Any(),
				testutils.EqVal(keyName),
				testutils.EqVal(version),
				gomock.Eq(encrypted),
			).
			Return(
				accessToken, // Decrypted access token
				nil,         // Error
			).
			MaxTimes(1)

		result, err := repo.Read(ctx, secret.SecretId)
		assert.NoError(t, err, "should be able to store the secret successfully")
		assert.Equal(t, secret.Value, result.Value, "should read the same value back")
	})

	t.Run("handle change", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t)
		kms := secrets.NewPlaintextKMS()
		ctx := context.Background()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		accessToken := gofakeit.UUID()

		repo := repository.NewSecretsRepository(log, clock, db, kms, user.AccountId)

		secret := repository.SecretData{
			Kind:  models.SecretKindPlaid,
			Value: accessToken,
		}
		err := repo.Store(ctx, &secret)
		assert.NoError(t, err, "should be able to store the secret successfully")
		assert.NotZero(t, secret.SecretId, "secret Id should now be set")

		result, err := repo.Read(ctx, secret.SecretId)
		assert.NoError(t, err, "should be able to store the secret successfully")
		assert.Equal(t, secret.Value, result.Value, "should read the same value back")

		accessTokenUpdated := gofakeit.UUID()
		assert.NotEqual(t, accessToken, accessTokenUpdated, "make sure the new token does not match the old one")

		secret.Value = accessTokenUpdated
		err = repo.Store(ctx, &secret)
		assert.NoError(t, err, "must be able to update an existing access token")

		result, err = repo.Read(ctx, secret.SecretId)
		assert.NoError(t, err, "must retrieve the written token")
		assert.Equal(t, accessTokenUpdated, result.Value, "retrieved token must match the second one written")
	})
}
