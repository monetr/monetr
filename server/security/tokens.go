package security

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/benbjohnson/clock"
	"github.com/monetr/monetr/server/build"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Audience string

const (
	AuthenticatedAudience Audience = "authenticated"
	ResetPasswordAudience Audience = "resetPassword"
	VerifyEmailAudience   Audience = "verifyEmail"
)

type Claims struct {
	// CreatedAt represents the timestamp the token was created, if this field is provided by a caller it will be
	// overwritten.
	CreatedAt    time.Time `json:"createdAt"`
	EmailAddress string    `json:"string"`
	UserId       uint64    `json:"userId,string"`
	AccountId    uint64    `json:"accountId,string"`
	LoginId      uint64    `json:"loginId,string"`
}

type ClientTokens interface {
	Parse(audience Audience, token string) (*Claims, error)
	Create(audience Audience, lifetime time.Duration, claims Claims) (string, error)
}

// In order to generate the public and private keys you will need to do this:
//
// openssl genpkey -algorithm ED25519 -out private.pem
// openssl pkey -in private.pem -out public.pem -pubout
//
// These can then be loaded into the app.
type pasetoClientTokens struct {
	log        *logrus.Entry
	clock      clock.Clock
	issuer     string
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	pub        paseto.V4AsymmetricPublicKey
	pri        paseto.V4AsymmetricSecretKey
}

func NewPasetoClientTokens(log *logrus.Entry, clock clock.Clock, issuer string, public ed25519.PublicKey, private ed25519.PrivateKey) (ClientTokens, error) {
	pub, err := paseto.NewV4AsymmetricPublicKeyFromEd25519(public)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create asymmetric public key from ed25519 provided")
	}

	pri, err := paseto.NewV4AsymmetricSecretKeyFromEd25519(private)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create asymmetric private key from ed25519 provided")
	}

	return &pasetoClientTokens{
		log:        log,
		clock:      clock,
		issuer:     issuer,
		publicKey:  public,
		privateKey: private,
		pub:        pub,
		pri:        pri,
	}, nil
}

func (p *pasetoClientTokens) Parse(audience Audience, token string) (*Claims, error) {
	parser := paseto.NewParserWithoutExpiryCheck()
	parser.AddRule(paseto.ForAudience(string(audience)))
	parser.AddRule(paseto.IssuedBy(p.issuer))
	parser.AddRule(p.notExpired())
	parser.AddRule(paseto.ValidAt(p.clock.Now()))

	parsedToken, err := parser.ParseV4Public(p.pub, token, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse token")
	}

	var claims Claims
	if err := parsedToken.Get("claims", &claims); err != nil {
		return nil, errors.Wrap(err, "failed to parse claims")
	}

	return &claims, nil
}

func (p *pasetoClientTokens) notExpired() paseto.Rule {
	return func(token paseto.Token) error {
		exp, err := token.GetExpiration()
		if err != nil {
			return err
		}

		if p.clock.Now().After(exp) {
			return errors.New("this token has expired")
		}

		return nil
	}
}

func (p *pasetoClientTokens) Create(audience Audience, lifetime time.Duration, claims Claims) (string, error) {
	token := paseto.NewToken()
	now := p.clock.Now()
	token.SetAudience(string(audience))
	token.SetExpiration(now.Add(lifetime))
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetIssuer(p.issuer)
	c := claims
	c.CreatedAt = p.clock.Now()
	err := token.Set("claims", c)
	if err != nil {
		panic(err)
	}
	token.SetSubject(c.EmailAddress)
	token.SetFooter([]byte(fmt.Sprintf("monetr %s - %s", build.Release, build.Revision)))

	result := token.V4Sign(p.pri, nil)
	return result, nil
}
