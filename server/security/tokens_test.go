package security_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"math"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/security"
	"github.com/stretchr/testify/assert"
)

func TestPasetoClientTokens(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)

		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		assert.NoError(t, err, "must be able to generate keys")

		clientTokens, err := security.NewPasetoClientTokens(log, clock, "monetr.local", publicKey, privateKey)
		assert.NoError(t, err, "must be able to init the client tokens interface")

		token, err := clientTokens.Create(security.AuthenticatedAudience, 5*time.Second, security.Claims{
			EmailAddress: gofakeit.Email(),
			UserId:       1,
			AccountId:    2,
			LoginId:      3,
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		claims, err := clientTokens.Parse(security.AuthenticatedAudience, token)
		assert.NoError(t, err, "should be able to parse the token it just generated")
		assert.NotNil(t, claims, "parsed token should not be nil")
		assert.EqualValues(t, 1, claims.UserId, "user Id should match expected")
		assert.EqualValues(t, 2, claims.AccountId, "account Id should match expected")
		assert.EqualValues(t, 3, claims.LoginId, "login Id should match expected")
	})

	t.Run("big Ids", func(t *testing.T) {
		// This is testing against a bug in the token library im using; it validates that the work around is working.
		var userId uint64 = math.MaxUint64
		var accountId uint64 = math.MaxUint64
		var loginId uint64 = math.MaxUint64

		clock := clock.NewMock()
		log := testutils.GetLog(t)

		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		assert.NoError(t, err, "must be able to generate keys")

		clientTokens, err := security.NewPasetoClientTokens(log, clock, "monetr.local", publicKey, privateKey)
		assert.NoError(t, err, "must be able to init the client tokens interface")

		token, err := clientTokens.Create(security.AuthenticatedAudience, 5*time.Second, security.Claims{
			EmailAddress: gofakeit.Email(),
			UserId:       userId,
			AccountId:    accountId,
			LoginId:      loginId,
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		claims, err := clientTokens.Parse(security.AuthenticatedAudience, token)
		assert.NoError(t, err, "should be able to parse the token it just generated")
		assert.NotNil(t, claims, "parsed token should not be nil")
		assert.EqualValues(t, userId, claims.UserId, "user Id should match expected")
		assert.EqualValues(t, accountId, claims.AccountId, "account Id should match expected")
		assert.EqualValues(t, loginId, claims.LoginId, "login Id should match expected")
	})

	t.Run("token expires", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)

		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		assert.NoError(t, err, "must be able to generate keys")

		clientTokens, err := security.NewPasetoClientTokens(log, clock, "monetr.local", publicKey, privateKey)
		assert.NoError(t, err, "must be able to init the client tokens interface")

		token, err := clientTokens.Create(security.AuthenticatedAudience, 5*time.Second, security.Claims{
			EmailAddress: gofakeit.Email(),
			UserId:       1,
			AccountId:    2,
			LoginId:      3,
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		clock.Add(10 * time.Second)

		claims, err := clientTokens.Parse(security.AuthenticatedAudience, token)
		assert.EqualError(t, err, "failed to parse token: this token has expired")
		assert.Nil(t, claims, "token should be nil when expired")
	})

	t.Run("different audience", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)

		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		assert.NoError(t, err, "must be able to generate keys")

		clientTokens, err := security.NewPasetoClientTokens(log, clock, "monetr.local", publicKey, privateKey)
		assert.NoError(t, err, "must be able to init the client tokens interface")

		token, err := clientTokens.Create(security.VerifyEmailAudience, 5*time.Second, security.Claims{
			EmailAddress: gofakeit.Email(),
			UserId:       1,
			AccountId:    2,
			LoginId:      3,
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		claims, err := clientTokens.Parse(security.AuthenticatedAudience, token)
		assert.EqualError(t, err, "failed to parse token: this token is not intended for `authenticated'. `verifyEmail' found")
		assert.Nil(t, claims, "token should be nil when expired")
	})

	t.Run("different keys", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)

		publicKey1, privateKey1, err := ed25519.GenerateKey(rand.Reader)
		assert.NoError(t, err, "must be able to generate keys")

		publicKey2, privateKey2, err := ed25519.GenerateKey(rand.Reader)
		assert.NoError(t, err, "must be able to generate keys")

		clientTokens1, err := security.NewPasetoClientTokens(log, clock, "monetr.local", publicKey1, privateKey1)
		assert.NoError(t, err, "must be able to init the client tokens interface")

		clientTokens2, err := security.NewPasetoClientTokens(log, clock, "monetr.local", publicKey2, privateKey2)
		assert.NoError(t, err, "must be able to init the client tokens interface")

		token, err := clientTokens1.Create(security.AuthenticatedAudience, 5*time.Second, security.Claims{
			EmailAddress: gofakeit.Email(),
			UserId:       1,
			AccountId:    2,
			LoginId:      3,
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		claims, err := clientTokens2.Parse(security.AuthenticatedAudience, token)
		assert.EqualError(t, err, "failed to parse token: bad signature")
		assert.Nil(t, claims, "token should be nil when it is not valid")
	})
}
