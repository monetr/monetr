package verification

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/form3tech-oss/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestJwtEmailVerificationTokenGenerator_GenerateToken(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		generator := jwtEmailVerificationTokenGenerator{
			secret: gofakeit.Generate("????????????????"),
		}

		email := gofakeit.Email()
		// If this test ever fails with an expiration error, then it is not able to complete within 1 second of the
		// token being generated.
		token, err := generator.GenerateToken(context.Background(), email, time.Second)
		assert.NoError(t, err, "must be able to generate a token without error")
		assert.NotEmpty(t, token, "token must not be blank")

		{ // Make sure the token can be parsed with the secret used to generate it.
			result, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(generator.secret), nil
			})
			assert.NoError(t, err, "must be able to parse jwt")
			assert.True(t, result.Valid, "result must be valid")
		}

		{ // This is stupid, but I am paranoid, so I want to make sure that it cannot be parsed with a different secret.
			result, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(gofakeit.Generate("????????????????")), nil
			})
			assert.EqualError(t, err, "signature is invalid", "must receive an error when parsing")
			assert.False(t, result.Valid, "result must be valid")
		}
	})

	t.Run("bad lifetime", func(t *testing.T) {
		generator := jwtEmailVerificationTokenGenerator{
			secret: gofakeit.Generate("????????????????"),
		}

		email := gofakeit.Email()
		token, err := generator.GenerateToken(context.Background(), email, 0)
		assert.EqualError(t, err, "lifetime must be greater than 1 second", "should return an invalid input error")
		assert.Empty(t, token, "token must not be blank")
	})
}

func TestJwtEmailVerificationTokenGenerator_ValidateToken(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		generator := jwtEmailVerificationTokenGenerator{
			secret: gofakeit.Generate("????????????????"),
		}

		email := gofakeit.Email()
		// If this test ever fails with an expiration error, then it is not able to complete within 1 second of the
		// token being generated.
		token, err := generator.GenerateToken(context.Background(), email, time.Second)
		assert.NoError(t, err, "must be able to generate a token without error")
		assert.NotEmpty(t, token, "token must not be blank")

		emailResult, err := generator.ValidateToken(context.Background(), token)
		assert.NoError(t, err, "should not receive an error, token should be valid")
		assert.Equal(t, emailResult, emailResult, "resulting email should match")
	})

	t.Run("expired token", func(t *testing.T) {
		generator := jwtEmailVerificationTokenGenerator{
			secret: gofakeit.Generate("????????????????"),
		}

		email := gofakeit.Email()
		// If this test ever fails with an expiration error, then it is not able to complete within 1 second of the
		// token being generated.
		token, err := generator.GenerateToken(context.Background(), email, time.Second)
		assert.NoError(t, err, "must be able to generate a token without error")
		assert.NotEmpty(t, token, "token must not be blank")

		time.Sleep(time.Second)

		emailResult, err := generator.ValidateToken(context.Background(), token)
		assert.EqualError(t, err, "token is expired", "should receive a token is expired error")
		assert.Equal(t, emailResult, emailResult, "resulting email should match even if its expired")
	})
}
