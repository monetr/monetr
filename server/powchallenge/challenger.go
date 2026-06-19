package powchallenge

import (
	"context"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"log/slog"
	"math/bits"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/getsentry/sentry-go"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/crumbs"
	"github.com/monetr/monetr/server/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// Purpose binds a challenge to one endpoint and is part of the signed wire
// format, so the values must not change.
type Purpose uint8

const (
	PurposeRegister Purpose = 1
	PurposeLogin    Purpose = 2
	PurposeForgot   Purpose = 3
	PurposeResend   Purpose = 4 // Resend verification email.
)

func (p Purpose) valid() bool {
	switch p {
	case PurposeRegister, PurposeLogin, PurposeForgot, PurposeResend:
		return true
	default:
		return false
	}
}

// Returned by Verify so the HTTP layer can map each failure to a specific (and
// not too revealing) client message.
var (
	ErrInvalidChallenge = errors.New("proof of work challenge is not valid")
	ErrChallengeExpired = errors.New("proof of work challenge has expired")
	ErrInvalidProof     = errors.New("proof of work solution is not valid")
	ErrChallengeReplay  = errors.New("proof of work challenge has already been used")
)

const (
	tokenVersion byte = 0x01
	// Versioned so a bot operator has to run our javascript, not a generic
	// solver.
	proofPrefix = "monetr-pow-v1:"
	// HKDF info string; HKDF is one way so this stays independent of the signing
	// key.
	secretInfo = "monetr-pow-secret-v1"
	// Clock skew tolerated on expiry, also padded onto the marker TTL.
	skewLeeway = 30 * time.Second

	// Token layout, then base32 (no padding) encoded for transport:
	//   [1]  version
	//   [1]  purpose
	//   [16] random
	//   [8]  issued at, unix seconds, big-endian
	//   [2]  difficulty, big-endian
	//   [32] HMAC-SHA256 of the preceding 28 bytes
	randomLength = 16
	signedLength = 28 // All but the trailing HMAC.
	tokenLength  = 60

	versionOffset    = 0
	purposeOffset    = 1
	randomOffset     = 2
	issuedAtOffset   = 18
	difficultyOffset = 26
	macOffset        = 28
)

// base32 keeps the token on a copy-paste safe alphabet (no +/ or -_).
var tokenEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// EncodedTokenLength is how many characters a challenge token is once it has
// been base32 encoded for the wire. The token is a fixed size so every
// challenge we issue is exactly this long. This is exported so the request
// validation can cheaply reject anything that is obviously not one of our
// tokens before we do the real (and authoritative) verification in Verify.
var EncodedTokenLength = tokenEncoding.EncodedLen(tokenLength)

type Challenge struct {
	Token      string `json:"challenge"`
	Difficulty int    `json:"difficulty"`
	// TTL in seconds, lets the client refetch a challenge that went stale while
	// idle.
	TTL int `json:"ttl"`
}

// Challenger issues and verifies the proof of work challenges that gate the
// unauthenticated authentication endpoints. See [config.ProofOfWork].
type Challenger interface {
	// Issue creates a challenge for the purpose and records a single-use marker.
	Issue(ctx context.Context, purpose Purpose) (*Challenge, error)
	// Verify checks the token and nonce, returns a typed error or nil, and
	// consumes the challenge on success so it cannot be reused.
	Verify(ctx context.Context, purpose Purpose, token string, nonce uint64) error
}

var (
	_ Challenger = &challengerBase{}
)

type challengerBase struct {
	log        *slog.Logger
	cache      cache.Cache
	clock      clock.Clock
	stats      *metrics.Stats
	secret     []byte
	difficulty int
	lifetime   time.Duration
}

func NewChallenger(
	log *slog.Logger,
	cache cache.Cache,
	clk clock.Clock,
	stats *metrics.Stats,
	secret []byte,
	difficulty int,
	lifetime time.Duration,
) Challenger {
	return &challengerBase{
		log:        log,
		cache:      cache,
		clock:      clk,
		stats:      stats,
		secret:     secret,
		difficulty: difficulty,
		lifetime:   lifetime,
	}
}

// DeriveSecret derives the 32 byte HMAC secret from the ED25519 seed via HKDF,
// which is one way so it stays independent of the signing key.
func DeriveSecret(seed []byte) []byte {
	secret, err := hkdf.Key(sha256.New, seed, nil, secretInfo, 32)
	if err != nil {
		// Only fails if the length is too large for the hash, which 32 bytes is
		// not.
		panic(errors.Wrap(err, "failed to derive proof of work secret"))
	}
	return secret
}

func (c *challengerBase) Issue(
	ctx context.Context,
	purpose Purpose,
) (*Challenge, error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	if !purpose.valid() {
		span.Status = sentry.SpanStatusInvalidArgument
		return nil, errors.Errorf("cannot issue a proof of work challenge for an unknown purpose: %d", purpose)
	}

	random := make([]byte, randomLength)
	if _, err := rand.Read(random); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to generate random bytes for proof of work challenge")
	}

	issuedAt := c.clock.Now().UTC()

	// Build and sign the token. Signing purpose locks it to one endpoint; signing
	// difficulty blocks reusing cheap tokens after a difficulty bump.
	token := make([]byte, tokenLength)
	token[versionOffset] = tokenVersion
	token[purposeOffset] = byte(purpose)
	copy(token[randomOffset:randomOffset+randomLength], random)
	binary.BigEndian.PutUint64(token[issuedAtOffset:issuedAtOffset+8], uint64(issuedAt.Unix()))
	binary.BigEndian.PutUint16(token[difficultyOffset:difficultyOffset+2], uint16(c.difficulty))
	copy(token[macOffset:], c.sign(token[:signedLength]))

	encoded := tokenEncoding.EncodeToString(token)

	// Plant the single-use marker, flipped to consumed in Verify. Key and values
	// are HMAC-derived so a redis-only attacker cannot forge one. The marker TTL
	// uses the wall clock; Verify's c.clock time check is the real expiry gate.
	if err := c.cache.SetTTL(
		span.Context(),
		c.replayKey(random),
		c.unusedValue(random),
		c.lifetime+skewLeeway,
	); err != nil {
		span.Status = sentry.SpanStatusInternalError
		return nil, errors.Wrap(err, "failed to persist proof of work challenge marker")
	}

	if c.stats != nil {
		c.stats.PowIssued.With(prometheus.Labels{}).Inc()
	}

	c.log.DebugContext(span.Context(), "issued proof of work challenge", "purpose", purpose, "difficulty", c.difficulty)

	span.Status = sentry.SpanStatusOK
	return &Challenge{
		Token:      encoded,
		Difficulty: c.difficulty,
		TTL:        int(c.lifetime.Seconds()),
	}, nil
}

func (c *challengerBase) Verify(
	ctx context.Context,
	expectedPurpose Purpose,
	token string,
	nonce uint64,
) (returnErr error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Record an outcome-labeled metric for every verification.
	start := c.clock.Now()
	result := "ok"
	defer func() {
		c.recordVerify(result, c.clock.Now().Sub(start))
	}()

	raw, err := tokenEncoding.DecodeString(token)
	if err != nil || len(raw) != tokenLength || raw[versionOffset] != tokenVersion {
		span.Status = sentry.SpanStatusInvalidArgument
		result = "invalid"
		return errors.WithStack(ErrInvalidChallenge)
	}

	// Check the HMAC before trusting any field; constant time to avoid leaking it.
	if subtle.ConstantTimeCompare(raw[macOffset:], c.sign(raw[:signedLength])) != 1 {
		span.Status = sentry.SpanStatusInvalidArgument
		result = "invalid"
		return errors.WithStack(ErrInvalidChallenge)
	}

	// Bound purpose must match the endpoint (no cross-endpoint reuse).
	if Purpose(raw[purposeOffset]) != expectedPurpose {
		span.Status = sentry.SpanStatusInvalidArgument
		result = "invalid"
		return errors.WithStack(ErrInvalidChallenge)
	}

	// Reject tokens minted below the current difficulty policy.
	difficulty := int(binary.BigEndian.Uint16(raw[difficultyOffset : difficultyOffset+2]))
	if difficulty < c.difficulty {
		span.Status = sentry.SpanStatusInvalidArgument
		result = "invalid"
		return errors.WithStack(ErrInvalidChallenge)
	}

	// Reject if expired; leeway tolerates a slightly-ahead issuing clock.
	issuedAt := time.Unix(int64(binary.BigEndian.Uint64(raw[issuedAtOffset:issuedAtOffset+8])), 0)
	age := c.clock.Now().UTC().Sub(issuedAt)
	if age > c.lifetime || age < -skewLeeway {
		span.Status = sentry.SpanStatusDeadlineExceeded
		result = "expired"
		return errors.WithStack(ErrChallengeExpired)
	}

	if leadingZeroBits(computeProofDigest(token, nonce)) < difficulty {
		span.Status = sentry.SpanStatusInvalidArgument
		result = "insufficient"
		return errors.WithStack(ErrInvalidProof)
	}

	// Atomically flip the marker unused -> consumed. No swap means used or gone,
	// so reject either way.
	random := raw[randomOffset : randomOffset+randomLength]
	swapped, err := c.cache.CompareAndSwap(
		span.Context(),
		c.replayKey(random),
		c.unusedValue(random),
		c.consumedValue(random),
		c.lifetime+skewLeeway,
	)
	if err != nil {
		// Fail closed: reject rather than risk a replay when redis is down.
		span.Status = sentry.SpanStatusInternalError
		result = "error"
		return errors.Wrap(err, "failed to verify proof of work challenge has not already been used")
	}
	if !swapped {
		span.Status = sentry.SpanStatusAlreadyExists
		result = "replay"
		return errors.WithStack(ErrChallengeReplay)
	}

	span.Status = sentry.SpanStatusOK
	return nil
}

// recordVerify emits the verification metrics; stats is nil in many tests.
func (c *challengerBase) recordVerify(result string, duration time.Duration) {
	if c.stats == nil {
		return
	}
	c.stats.PowVerified.With(prometheus.Labels{
		"result": result,
	}).Inc()
	c.stats.PowVerifyTime.With(prometheus.Labels{
		"result": result,
	}).Observe(float64(duration.Milliseconds()))
}

func (c *challengerBase) sign(data []byte) []byte {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write(data)
	return mac.Sum(nil)
}

// Cache key for the marker, HMAC'd so it is not predictable without the secret.
func (c *challengerBase) replayKey(random []byte) string {
	return "pow:" + tokenEncoding.EncodeToString(c.tag("pow-key:", random))
}

func (c *challengerBase) unusedValue(random []byte) []byte {
	return c.tag("pow-unused:", random)
}

func (c *challengerBase) consumedValue(random []byte) []byte {
	return c.tag("pow-consumed:", random)
}

// HMAC a label with the random; distinct labels keep key/unused/consumed different.
func (c *challengerBase) tag(label string, random []byte) []byte {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(label))
	mac.Write(random)
	return mac.Sum(nil)
}

// The exact bytes the client hashes: prefix + token + ":" + 8 big-endian nonce
// bytes. Must match the frontend solver (the shared test vectors guard drift).
func computeProofDigest(token string, nonce uint64) []byte {
	h := sha256.New()
	h.Write([]byte(proofPrefix))
	h.Write([]byte(token))
	h.Write([]byte(":"))
	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], nonce)
	h.Write(nonceBytes[:])
	return h.Sum(nil)
}

// Count leading zero bits (not bytes or hex digits) so difficulty tunes one bit
// at a time.
func leadingZeroBits(digest []byte) int {
	var count int
	for _, b := range digest {
		if b == 0 {
			count += 8
			continue
		}
		count += bits.LeadingZeros8(b)
		break
	}
	return count
}
