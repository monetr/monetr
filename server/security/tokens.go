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

type Scope string

const (
	AuthenticatedScope Scope = "authenticated"
	MultiFactorScope   Scope = "multiFactor"
	ResetPasswordScope Scope = "resetPassword"
	VerifyEmailScope   Scope = "verifyEmail"
)

type Claims struct {
	// CreatedAt represents the timestamp the token was created, if this field is
	// provided by a caller it will be overwritten.
	CreatedAt    time.Time `json:"createdAt"`
	EmailAddress string    `json:"email,string"`
	UserId       string    `json:"userId,string"`
	AccountId    string    `json:"accountId,string"`
	LoginId      string    `json:"loginId,string"`
	Scope        Scope     `json:"scope,string"`
	// ReissueCount will be used to allow tokens to be reissued a certain number
	// of times to allow users to stay logged in for a longer period of time if
	// they are using the application frequently. Tokens will not be reissued if
	// they have expired.
	ReissueCount uint8 `json:"reissueCount,string"`
}

// RequireScope takes an array of allowed scopes. If the claim is any one of the
// specified scopes then this function will return nil. If the claim does not
// contain any of the specified scopes then this will return an error.
func (c Claims) RequireScope(scopes ...Scope) error {
	if c.Scope == "" {
		return errors.New("authentication is missing scope")
	}

	for i := range scopes {
		scope := scopes[i]
		// If the claims have one of the scopes then just return nil. The claims are
		// valid.
		if scope == c.Scope {
			return nil
		}
	}

	return errors.Errorf("authentication does not have required scope; has: [%s] required: %v", c.Scope, scopes)
}

func (c Claims) Valid() error {
	if c.CreatedAt.IsZero() {
		return errors.New("claims invalid: created at is zero")
	}

	if c.LoginId == "" {
		return errors.New("claims invalid: login Id is not defined")
	}

	if c.Scope == "" {
		return errors.New("claims invalid: scope is not defined")
	}

	if c.Scope == AuthenticatedScope {
		if c.AccountId == "" {
			return errors.New("claims invalid: account Id is not defined")
		}

		if c.UserId == "" {
			return errors.New("claims invalid: userId Id is not defined")
		}
	}

	return nil
}

type ClientTokens interface {
	Parse(token string) (*Claims, error)
	Create(lifetime time.Duration, claims Claims) (string, error)
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

func NewPasetoClientTokens(
	log *logrus.Entry,
	clock clock.Clock,
	issuer string,
	public ed25519.PublicKey,
	private ed25519.PrivateKey,
) (ClientTokens, error) {
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

func (p *pasetoClientTokens) Parse(token string) (*Claims, error) {
	parser := paseto.NewParserWithoutExpiryCheck()
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

func (p *pasetoClientTokens) Create(
	lifetime time.Duration,
	claims Claims,
) (string, error) {
	token := paseto.NewToken()
	now := p.clock.Now()
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
