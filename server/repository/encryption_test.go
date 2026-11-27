package repository_test

import (
	"database/sql"
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func givenIHaveARootKey(
	t *testing.T,
	repo repository.EncryptionRepository,
) *models.RootKey {
	version := "1"
	key := &models.RootKey{
		Wrappers: []models.RootKeyWrapper{
			{
				Kind:      models.RootKeyKindSecretBox,
				Reference: gofakeit.UUID(),
				Version:   &version,
				Data:      gofakeit.UUID(),
			},
		},
		Fingerprint: gofakeit.UUID(),
	}
	require.NoError(t, repo.CreateRootKey(t.Context(), key), "must create root key")
	return key
}

func givenIHaveAKeyEncryptionKey(
	t *testing.T,
	repo repository.EncryptionRepository,
	rootKey *models.RootKey,
) *models.KeyEncryptionKey {
	key := &models.KeyEncryptionKey{
		RootKeyId: rootKey.RootKeyId,
		Wrapped:   gofakeit.UUID(),
	}
	require.NoError(t, repo.CreateKeyEncryptionKey(t.Context(), key), "must create key encryption key")
	return key
}

func givenIHaveASecretboxKey(
	t *testing.T,
	repo repository.EncryptionRepository,
	kek *models.KeyEncryptionKey,
) *models.SecretboxKey {
	key := &models.SecretboxKey{
		Fingerprint: gofakeit.UUID(),
		Encrypted: models.Encrypted[string]{
			KeyEncryptionKeyId: kek.KeyEncryptionKeyId,
			EncryptedData:      gofakeit.UUID(),
		},
	}
	require.NoError(t, repo.CreateSecretboxKey(t.Context(), key), "must create secretbox key")
	return key
}

func TestEncryptionRepository_CreateRootKey(t *testing.T) {
	t.Run("stores the root key", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		key := givenIHaveARootKey(t, repo)
		assert.NotZero(t, key.RootKeyId, "root key ID should be generated")
		assert.NotZero(t, key.CreatedAt, "created at should be set")

		stored, err := repo.GetRootKeyById(t.Context(), key.RootKeyId)
		require.NoError(t, err, "must read the root key back")
		assert.Equal(t, key.RootKeyId, stored.RootKeyId)
		assert.Equal(t, key.Fingerprint, stored.Fingerprint)
		assert.Equal(t, key.Wrappers, stored.Wrappers, "wrappers should round trip through jsonb")
	})

	t.Run("duplicate fingerprint", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		existing := givenIHaveARootKey(t, repo)

		err := repo.CreateRootKey(t.Context(), &models.RootKey{
			Wrappers:    existing.Wrappers,
			Fingerprint: existing.Fingerprint,
		})
		assert.Error(t, err, "should not store two root keys with the same fingerprint")
		assert.Contains(t, err.Error(), "failed to store root key")
	})
}

func TestEncryptionRepository_CreateKeyEncryptionKey(t *testing.T) {
	t.Run("stores the key encryption key", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		rootKey := givenIHaveARootKey(t, repo)
		key := givenIHaveAKeyEncryptionKey(t, repo, rootKey)
		assert.NotZero(t, key.KeyEncryptionKeyId, "key encryption key ID should be generated")
		assert.NotZero(t, key.CreatedAt, "created at should be set")

		stored, err := repo.GetKeyEncryptionKeyById(t.Context(), key.KeyEncryptionKeyId)
		require.NoError(t, err, "must read the key encryption key back")
		assert.Equal(t, key.KeyEncryptionKeyId, stored.KeyEncryptionKeyId)
		assert.Equal(t, rootKey.RootKeyId, stored.RootKeyId)
		assert.Equal(t, key.Wrapped, stored.Wrapped)
	})

	t.Run("root key does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		err := repo.CreateKeyEncryptionKey(t.Context(), &models.KeyEncryptionKey{
			RootKeyId: models.NewID[models.RootKey](),
			Wrapped:   gofakeit.UUID(),
		})
		assert.Error(t, err, "should not store a key encryption key for a missing root key")
		assert.Contains(t, err.Error(), "failed to store key encryption key")
		assert.Contains(t, err.Error(), "SQLSTATE=23503", "should be a foreign key violation")
	})
}

func TestEncryptionRepository_CreateSecretboxKey(t *testing.T) {
	t.Run("stores the secretbox key", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		rootKey := givenIHaveARootKey(t, repo)
		kek := givenIHaveAKeyEncryptionKey(t, repo, rootKey)
		key := givenIHaveASecretboxKey(t, repo, kek)
		assert.NotZero(t, key.SecretboxKeyId, "secretbox key ID should be generated")
		assert.NotZero(t, key.CreatedAt, "created at should be set")

		stored, err := repo.GetSecretboxKey(t.Context())
		require.NoError(t, err, "must read the secretbox key back")
		assert.Equal(t, key.SecretboxKeyId, stored.SecretboxKeyId)
		assert.Equal(t, key.Fingerprint, stored.Fingerprint)
		assert.Equal(t, key.KeyEncryptionKeyId, stored.KeyEncryptionKeyId)
		assert.Equal(t, key.EncryptedData, stored.EncryptedData)
	})

	t.Run("key encryption key does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		err := repo.CreateSecretboxKey(t.Context(), &models.SecretboxKey{
			Fingerprint: gofakeit.UUID(),
			Encrypted: models.Encrypted[string]{
				KeyEncryptionKeyId: models.NewID[models.KeyEncryptionKey](),
				EncryptedData:      gofakeit.UUID(),
			},
		})
		assert.Error(t, err, "should not store a secretbox key for a missing key encryption key")
		assert.Contains(t, err.Error(), "failed to store secretbox key")
		assert.Contains(t, err.Error(), "SQLSTATE=23503", "should be a foreign key violation")
	})

	t.Run("not encrypted", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		err := repo.CreateSecretboxKey(t.Context(), &models.SecretboxKey{
			Fingerprint: gofakeit.UUID(),
		})
		assert.Error(t, err, "secretbox keys must always be stored encrypted")
		assert.Contains(t, err.Error(), "failed to store secretbox key")
		assert.Contains(t, err.Error(), "SQLSTATE=23502", "should be a not null violation")
	})
}

func TestEncryptionRepository_GetRootKey(t *testing.T) {
	t.Run("returns the highest root key", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		first := givenIHaveARootKey(t, repo)
		second := givenIHaveARootKey(t, repo)

		current, err := repo.GetRootKey(t.Context())
		require.NoError(t, err, "must read the current root key")
		assert.Equal(t, max(first.RootKeyId, second.RootKeyId), current.RootKeyId)
	})

	t.Run("no root keys", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		current, err := repo.GetRootKey(t.Context())
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, current)
	})
}

func TestEncryptionRepository_GetRootKeyById(t *testing.T) {
	t.Run("returns the requested root key", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		first := givenIHaveARootKey(t, repo)
		second := givenIHaveARootKey(t, repo)

		for _, key := range []*models.RootKey{first, second} {
			stored, err := repo.GetRootKeyById(t.Context(), key.RootKeyId)
			require.NoError(t, err, "must read the root key by ID")
			assert.Equal(t, key.RootKeyId, stored.RootKeyId)
			assert.Equal(t, key.Fingerprint, stored.Fingerprint)
		}
	})

	t.Run("does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		givenIHaveARootKey(t, repo)

		stored, err := repo.GetRootKeyById(t.Context(), models.NewID[models.RootKey]())
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, stored)
	})
}

func TestEncryptionRepository_GetKeyEncryptionKey(t *testing.T) {
	t.Run("returns the highest key encryption key", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		rootKey := givenIHaveARootKey(t, repo)
		first := givenIHaveAKeyEncryptionKey(t, repo, rootKey)
		second := givenIHaveAKeyEncryptionKey(t, repo, rootKey)

		current, err := repo.GetKeyEncryptionKey(t.Context())
		require.NoError(t, err, "must read the current key encryption key")
		assert.Equal(t, max(first.KeyEncryptionKeyId, second.KeyEncryptionKeyId), current.KeyEncryptionKeyId)
	})

	t.Run("no key encryption keys", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		current, err := repo.GetKeyEncryptionKey(t.Context())
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, current)
	})
}

func TestEncryptionRepository_GetKeyEncryptionKeyById(t *testing.T) {
	t.Run("returns the requested key encryption key", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		rootKey := givenIHaveARootKey(t, repo)
		first := givenIHaveAKeyEncryptionKey(t, repo, rootKey)
		second := givenIHaveAKeyEncryptionKey(t, repo, rootKey)

		for _, key := range []*models.KeyEncryptionKey{first, second} {
			stored, err := repo.GetKeyEncryptionKeyById(t.Context(), key.KeyEncryptionKeyId)
			require.NoError(t, err, "must read the key encryption key by ID")
			assert.Equal(t, key.KeyEncryptionKeyId, stored.KeyEncryptionKeyId)
			assert.Equal(t, key.Wrapped, stored.Wrapped)
		}
	})

	t.Run("does not exist", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		rootKey := givenIHaveARootKey(t, repo)
		givenIHaveAKeyEncryptionKey(t, repo, rootKey)

		stored, err := repo.GetKeyEncryptionKeyById(t.Context(), models.NewID[models.KeyEncryptionKey]())
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, stored)
	})
}

func TestEncryptionRepository_GetSecretboxKey(t *testing.T) {
	t.Run("returns the highest secretbox key", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		rootKey := givenIHaveARootKey(t, repo)
		kek := givenIHaveAKeyEncryptionKey(t, repo, rootKey)
		first := givenIHaveASecretboxKey(t, repo, kek)
		second := givenIHaveASecretboxKey(t, repo, kek)

		current, err := repo.GetSecretboxKey(t.Context())
		require.NoError(t, err, "must read the current secretbox key")
		assert.Equal(t, max(first.SecretboxKeyId, second.SecretboxKeyId), current.SecretboxKeyId)
	})

	t.Run("no secretbox keys", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)
		db := testutils.GetPgDatabase(t, testutils.IsolatedDatabase)
		repo := repository.NewEncryptionRepository(log, clock, db)

		current, err := repo.GetSecretboxKey(t.Context())
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, current)
	})
}
