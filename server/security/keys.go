package security

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base32"
	"strings"

	"github.com/monetr/monetr/server/models"
	"github.com/pkg/errors"
)

// GenerateKey takes a user model and generates an API key object for the user,
// returning the secret that is used for they key. This secret is not
// recoverable.
func GenerateKey(user models.User) (key *models.Key, secret string, err error) {
	secretBytes := make([]byte, 64)
	if _, err := rand.Read(secretBytes); err != nil {
		return nil, "", errors.WithStack(err)
	}

	sha512 := sha512.New()
	if _, err := sha512.Write(secretBytes); err != nil {
		return nil, "", errors.WithStack(err)
	}

	secret = strings.ToLower(base32.StdEncoding.EncodeToString(secretBytes))

	key = &models.Key{
		UserId:    user.UserId,
		User:      &user,
		AccountId: user.AccountId,
		Verifier:  sha512.Sum(nil),
	}

	return key, secret, nil
}
