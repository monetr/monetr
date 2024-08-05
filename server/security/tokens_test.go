package security_test

import (
	"crypto/ed25519"
	"crypto/rand"
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

		token, err := clientTokens.Create(5*time.Second, security.Claims{
			Scope:        security.AuthenticatedScope,
			EmailAddress: gofakeit.Email(),
			UserId:       "user_1",
			AccountId:    "acct_2",
			LoginId:      "lgn_3",
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		claims, err := clientTokens.Parse(token)
		assert.NoError(t, err, "should be able to parse the token it just generated")
		assert.NotNil(t, claims, "parsed token should not be nil")
		assert.EqualValues(t, "user_1", claims.UserId, "user Id should match expected")
		assert.EqualValues(t, "acct_2", claims.AccountId, "account Id should match expected")
		assert.EqualValues(t, "lgn_3", claims.LoginId, "login Id should match expected")
	})

	t.Run("token expires", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)

		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		assert.NoError(t, err, "must be able to generate keys")

		clientTokens, err := security.NewPasetoClientTokens(log, clock, "monetr.local", publicKey, privateKey)
		assert.NoError(t, err, "must be able to init the client tokens interface")

		token, err := clientTokens.Create(5*time.Second, security.Claims{
			Scope:        security.AuthenticatedScope,
			EmailAddress: gofakeit.Email(),
			UserId:       "user_1",
			AccountId:    "acct_2",
			LoginId:      "lgn_3",
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		clock.Add(10 * time.Second)

		claims, err := clientTokens.Parse(token)
		assert.EqualError(t, err, "failed to parse token: this token has expired")
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

		token, err := clientTokens1.Create(5*time.Second, security.Claims{
			Scope:        security.AuthenticatedScope,
			EmailAddress: gofakeit.Email(),
			UserId:       "user_1",
			AccountId:    "acct_2",
			LoginId:      "lgn_3",
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		claims, err := clientTokens2.Parse(token)
		assert.EqualError(t, err, "failed to parse token: bad signature")
		assert.Nil(t, claims, "token should be nil when it is not valid")
	})
}

func TestClaimsRequiredScope(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		clock := clock.NewMock()
		log := testutils.GetLog(t)

		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		assert.NoError(t, err, "must be able to generate keys")

		clientTokens, err := security.NewPasetoClientTokens(log, clock, "monetr.local", publicKey, privateKey)
		assert.NoError(t, err, "must be able to init the client tokens interface")

		token, err := clientTokens.Create(5*time.Second, security.Claims{
			Scope:        security.AuthenticatedScope,
			EmailAddress: gofakeit.Email(),
			UserId:       "user_1",
			AccountId:    "acct_2",
			LoginId:      "lgn_3",
		})
		assert.NoError(t, err, "must be able to create a token successfully")
		assert.NotEmpty(t, token, "token must not be empty")

		claims, err := clientTokens.Parse(token)
		assert.NoError(t, err, "should be able to parse token into claims")
		assert.NoError(t, claims.RequireScope(security.AuthenticatedScope), "claims should have the AuthenticatedAudience scope")
		assert.EqualError(t, claims.RequireScope(security.ResetPasswordScope), "authentication does not have required scope; has: [authenticated] required: [resetPassword]")
		assert.EqualError(t, claims.RequireScope(security.VerifyEmailScope), "authentication does not have required scope; has: [authenticated] required: [verifyEmail]")
		assert.EqualError(t, claims.RequireScope(security.VerifyEmailScope, security.ResetPasswordScope), "authentication does not have required scope; has: [authenticated] required: [verifyEmail resetPassword]")
	})
}
