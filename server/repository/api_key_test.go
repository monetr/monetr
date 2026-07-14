package repository_test

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/internal/fixtures"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/models"
	"github.com/monetr/monetr/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetTestRepositoryForUser(
	t *testing.T,
	clock clock.Clock,
	user models.User,
) repository.Repository {
	return repository.NewRepositoryFromSession(
		clock,
		user.UserId,
		user.AccountId,
		testutils.GetPgDatabase(t),
		testutils.GetLog(t),
	)
}

func GivenIHaveAnApiKey(
	t *testing.T,
	repo repository.Repository,
	name string,
) *models.ApiKey {
	key, _, err := models.NewApiKey()
	require.NoError(t, err, "must be able to generate a new api key")
	key.Name = name

	err = repo.CreateApiKey(t.Context(), key)
	require.NoError(t, err, "must be able to create the api key")

	return key
}

func TestRepositoryBase_CreateApiKey(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)

		key, _, err := models.NewApiKey()
		require.NoError(t, err, "must be able to generate a new api key")
		key.Name = "Test API Key"

		err = repo.CreateApiKey(t.Context(), key)
		require.NoError(t, err, "must be able to create the api key")

		assert.False(t, key.ApiKeyId.IsZero(), "an Id should have been assigned")
		assert.Equal(t, user.AccountId, key.AccountId, "the key should belong to the session's account")
		assert.Equal(t, user.UserId, key.CreatedBy, "the key should be created by the session's user")
		assert.False(t, key.CreatedAt.IsZero(), "created at should have been set")
		testutils.MustDBExist(t, *key)
	})
}

func TestRepositoryBase_GetApiKeys(t *testing.T) {
	t.Run("returns the account's keys", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)

		keys, err := repo.GetApiKeys(t.Context())
		require.NoError(t, err, "must be able to list api keys")
		assert.Empty(t, keys, "there should be no keys to begin with")

		GivenIHaveAnApiKey(t, repo, "First Key")
		GivenIHaveAnApiKey(t, repo, "Second Key")

		keys, err = repo.GetApiKeys(t.Context())
		require.NoError(t, err, "must be able to list api keys")
		assert.Len(t, keys, 2, "both keys should be returned")
	})

	t.Run("excludes revoked keys", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)

		revoked := GivenIHaveAnApiKey(t, repo, "Revoked Key")
		active := GivenIHaveAnApiKey(t, repo, "Active Key")

		err := repo.DeleteApiKey(t.Context(), revoked.ApiKeyId)
		require.NoError(t, err, "must be able to revoke the api key")

		keys, err := repo.GetApiKeys(t.Context())
		require.NoError(t, err, "must be able to list api keys")
		require.Len(t, keys, 1, "only the active key should be returned")
		assert.Equal(t, active.ApiKeyId, keys[0].ApiKeyId, "the returned key should be the active one")
	})

	t.Run("is scoped to the account", func(t *testing.T) {
		clock := clock.NewMock()
		userOne, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		userTwo, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repoOne := GetTestRepositoryForUser(t, clock, userOne)
		repoTwo := GetTestRepositoryForUser(t, clock, userTwo)

		GivenIHaveAnApiKey(t, repoOne, "Account One Key")

		keys, err := repoTwo.GetApiKeys(t.Context())
		require.NoError(t, err, "must be able to list api keys")
		assert.Empty(t, keys, "the second account must not see the first account's keys")
	})
}

func TestRepositoryBase_GetApiKeyById(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)

		key := GivenIHaveAnApiKey(t, repo, "Test API Key")

		fetched, err := repo.GetApiKeyById(t.Context(), key.ApiKeyId)
		require.NoError(t, err, "must be able to retrieve the api key by Id")
		require.NotNil(t, fetched, "the api key should have been returned")
		assert.Equal(t, key.ApiKeyId, fetched.ApiKeyId, "the retrieved key should match")

		// The creating user and their login must be eagerly loaded, the
		// authentication middleware relies on this.
		require.NotNil(t, fetched.CreatedByUser, "the created by user relation should be loaded")
		require.NotNil(t, fetched.CreatedByUser.Login, "the created by user's login relation should be loaded")
		assert.Equal(t, user.Login.Email, fetched.CreatedByUser.Login.Email, "the login email should match the creating user")
	})

	t.Run("returns revoked keys", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)

		key := GivenIHaveAnApiKey(t, repo, "Revoked Key")
		err := repo.DeleteApiKey(t.Context(), key.ApiKeyId)
		require.NoError(t, err, "must be able to revoke the api key")

		// GetApiKeyById does not filter on deleted_at, the controller relies on
		// being able to read a revoked key back in order to report that it is
		// already revoked.
		fetched, err := repo.GetApiKeyById(t.Context(), key.ApiKeyId)
		require.NoError(t, err, "must still be able to retrieve a revoked key by Id")
		require.NotNil(t, fetched, "the revoked key should have been returned")
		assert.NotNil(t, fetched.DeletedAt, "the retrieved key should be marked as revoked")
	})

	t.Run("cannot read another account's key", func(t *testing.T) {
		clock := clock.NewMock()
		userOne, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		userTwo, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repoOne := GetTestRepositoryForUser(t, clock, userOne)
		repoTwo := GetTestRepositoryForUser(t, clock, userTwo)

		key := GivenIHaveAnApiKey(t, repoOne, "Account One Key")

		fetched, err := repoTwo.GetApiKeyById(t.Context(), key.ApiKeyId)
		assert.Error(t, err, "must not be able to read another account's key")
		assert.Nil(t, fetched, "no key should be returned")
	})
}

func TestRepositoryBase_DeleteApiKey(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)

		key := GivenIHaveAnApiKey(t, repo, "Test API Key")

		err := repo.DeleteApiKey(t.Context(), key.ApiKeyId)
		require.NoError(t, err, "must be able to revoke the api key")

		revoked := testutils.MustDBRead(t, *key)
		assert.NotNil(t, revoked.DeletedAt, "deleted at should be set after revoking")
	})

	t.Run("cannot revoke a key twice", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)

		key := GivenIHaveAnApiKey(t, repo, "Test API Key")

		err := repo.DeleteApiKey(t.Context(), key.ApiKeyId)
		require.NoError(t, err, "the first revocation should succeed")

		err = repo.DeleteApiKey(t.Context(), key.ApiKeyId)
		assert.EqualError(t, err, "invalid api key specified or key is already deactivated", "revoking an already revoked key should fail")
	})

	t.Run("cannot revoke another account's key", func(t *testing.T) {
		clock := clock.NewMock()
		userOne, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		userTwo, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repoOne := GetTestRepositoryForUser(t, clock, userOne)
		repoTwo := GetTestRepositoryForUser(t, clock, userTwo)

		key := GivenIHaveAnApiKey(t, repoOne, "Account One Key")

		err := repoTwo.DeleteApiKey(t.Context(), key.ApiKeyId)
		assert.Error(t, err, "must not be able to revoke another account's key")

		// The key must still be active for its own account.
		stillActive := testutils.MustDBRead(t, *key)
		assert.Nil(t, stillActive.DeletedAt, "the key must not have been revoked")
	})
}

func TestUnauthenticatedRepo_GetApiKey(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)
		unauthenticated := GetTestUnauthenticatedRepository(t, clock)

		key := GivenIHaveAnApiKey(t, repo, "Test API Key")

		fetched, err := unauthenticated.GetApiKey(t.Context(), key.ApiKeyId)
		require.NoError(t, err, "must be able to retrieve the api key")
		require.NotNil(t, fetched, "the api key should have been returned")
		assert.Equal(t, key.ApiKeyId, fetched.ApiKeyId, "the retrieved key should match")
		assert.Equal(t, user.AccountId, fetched.AccountId, "the key should carry its owning account")

		// The authentication middleware builds its claims from these relations, so
		// they must be populated.
		require.NotNil(t, fetched.CreatedByUser, "the created by user relation should be loaded")
		require.NotNil(t, fetched.CreatedByUser.Login, "the created by user's login relation should be loaded")
		assert.Equal(t, user.Login.Email, fetched.CreatedByUser.Login.Email, "the login email should match the creating user")
	})

	t.Run("excludes revoked keys", func(t *testing.T) {
		clock := clock.NewMock()
		user, _ := fixtures.GivenIHaveABasicAccount(t, clock)
		repo := GetTestRepositoryForUser(t, clock, user)
		unauthenticated := GetTestUnauthenticatedRepository(t, clock)

		key := GivenIHaveAnApiKey(t, repo, "Revoked Key")
		err := repo.DeleteApiKey(t.Context(), key.ApiKeyId)
		require.NoError(t, err, "must be able to revoke the api key")

		fetched, err := unauthenticated.GetApiKey(t.Context(), key.ApiKeyId)
		assert.Error(t, err, "a revoked key must not be retrievable for authentication")
		assert.Nil(t, fetched, "no key should be returned")
	})

	t.Run("unknown key", func(t *testing.T) {
		clock := clock.NewMock()
		unauthenticated := GetTestUnauthenticatedRepository(t, clock)

		fetched, err := unauthenticated.GetApiKey(t.Context(), models.NewID[models.ApiKey]())
		assert.Error(t, err, "an unknown key must return an error")
		assert.Nil(t, fetched, "no key should be returned")
	})
}
