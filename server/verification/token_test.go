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
		assert.Equal(t, email, emailResult, "resulting email should match")
	})

	t.Run("expired token", func(t *testing.T) {
		generator := jwtEmailVerificationTokenGenerator{
			secret: gofakeit.Generate("????????????????"),
		}

		email := gofakeit.Email()
		token, err := generator.GenerateToken(context.Background(), email, time.Second)
		assert.NoError(t, err, "must be able to generate a token without error")
		assert.NotEmpty(t, token, "token must not be blank")

		time.Sleep(2 * time.Second)

		emailResult, err := generator.ValidateToken(context.Background(), token)
		assert.Regexp(t, `invalid token: token is expired by \ds`, err.Error(), "should receive a token is expired error")
		assert.Empty(t, emailResult, "email should not be returned if the token is expired")
	})

	t.Run("invalid token", func(t *testing.T) {
		email := gofakeit.Email()

		var badToken string
		{ // Generate a token using one secret.
			generator := jwtEmailVerificationTokenGenerator{
				secret: gofakeit.Generate("????????????????"),
			}
			token, err := generator.GenerateToken(context.Background(), email, time.Minute)
			assert.NoError(t, err, "must be able to generate a token without error")
			assert.NotEmpty(t, token, "token must not be blank")

			badToken = token
		}

		{ // Then try to validate the token with a different secret.
			generator := jwtEmailVerificationTokenGenerator{
				secret: gofakeit.Generate("????????????????"),
			}
			emailResult, err := generator.ValidateToken(context.Background(), badToken)
			assert.EqualError(t, err, "invalid token: signature is invalid", "should receive a token is invalid error regarding the signature")
			assert.Equal(t, emailResult, emailResult, "resulting email should match even if its expired")
		}
	})
}
