package verification

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

type TokenValidation struct {
	Email     string
	CreatedAt time.Time
}

type TokenGenerator interface {
	// GenerateToken will take an email address and a lifespan for the resulting token. It will return a string of
	// arbitrary characters that can be passed to ValidateToken. The lifetime must be positive and the email address
	// cannot be blank.
	GenerateToken(ctx context.Context, emailAddress string, lifetime time.Duration) (token string, _ error)
	// ValidateToken receives a token string and will make sure that it is usable. If it is it will return the email
	// address for that token. If it is not it will return an error indicating that the token is not valid.
	ValidateToken(ctx context.Context, token string) (emailAddress string, _ error)

	ValidateTokenEx(ctx context.Context, token string) (*TokenValidation, error)
}

var (
	_ TokenGenerator = &jwtEmailVerificationTokenGenerator{}
)

type jwtEmailVerificationTokenGenerator struct {
	secret string
}

func NewTokenGenerator(secret string) TokenGenerator {
	return jwtEmailVerificationTokenGenerator{
		secret: secret,
	}
}

func (j jwtEmailVerificationTokenGenerator) GenerateToken(ctx context.Context, emailAddress string, lifetime time.Duration) (token string, err error) {
	span := sentry.StartSpan(ctx, "JWTVerification - GenerateToken")
	defer span.Finish()

	if lifetime < time.Second {
		return token, errors.New("lifetime must be greater than 1 second")
	}

	emailAddress = strings.ToLower(strings.TrimSpace(emailAddress))
	if emailAddress == "" {
		return token, errors.New("email address cannot be empty or whitespace")
	}

	checksum := md5.Sum([]byte(emailAddress))
	id := fmt.Sprintf("%X", string(checksum[:]))

	now := time.Now()
	claims := jwt.StandardClaims{
		Audience: []string{
			emailAddress,
		},
		ExpiresAt: now.Add(lifetime).Unix(),
		Id:        id,
		IssuedAt:  now.Unix(),
		Issuer:    "monetr",
		NotBefore: now.Unix(),
		Subject:   "monetr.email.verification",
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = jwtToken.SignedString([]byte(j.secret))
	return
}

func (j jwtEmailVerificationTokenGenerator) ValidateToken(ctx context.Context, token string) (emailAddress string, _ error) {
	result, err := j.ValidateTokenEx(ctx, token)
	if err != nil {
		return emailAddress, err
	}

	return result.Email, nil
}

func (j jwtEmailVerificationTokenGenerator) ValidateTokenEx(ctx context.Context, token string) (*TokenValidation, error) {
	span := sentry.StartSpan(ctx, "JWTVerification - ValidateToken")
	defer span.Finish()

	var emailAddress string
	var claims jwt.StandardClaims
	result, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "invalid token")
	}

	if !result.Valid {
		return nil, errors.New("invalid token")
	}

	if len(claims.Audience) != 1 {
		return nil, errors.New("invalid audience on token")
	}

	emailAddress = claims.Audience[0]

	checksum := md5.Sum([]byte(emailAddress))
	id := fmt.Sprintf("%X", string(checksum[:]))
	if claims.Id != id {
		return nil, errors.New("token ID is not valid")
	}

	return &TokenValidation{
		Email:     emailAddress,
		CreatedAt: time.Unix(claims.IssuedAt, 0),
	}, nil
}
