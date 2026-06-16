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

// Purpose binds a challenge to a single endpoint, it is baked into the signed
// token. The numeric values are part of the wire format and must not change.
type Purpose uint8

const (
	PurposeRegister Purpose = 1
	PurposeLogin    Purpose = 2
	PurposeForgot   Purpose = 3
)

func (p Purpose) valid() bool {
	switch p {
	case PurposeRegister, PurposeLogin, PurposeForgot:
		return true
	default:
		return false
	}
}

// These are returned by [Challenger.Verify] so the HTTP layer can map each
// failure to a specific (and not too revealing) message for the client.
var (
	ErrInvalidChallenge = errors.New("proof of work challenge is not valid")
	ErrChallengeExpired = errors.New("proof of work challenge has expired")
	ErrInvalidProof     = errors.New("proof of work solution is not valid")
	ErrChallengeReplay  = errors.New("proof of work challenge has already been used")
)

const (
	tokenVersion byte = 0x01
	// proofPrefix is mixed into the hash. Versioning it forces a bot operator to
	// run our javascript instead of reusing a generic solver.
	proofPrefix = "monetr-pow-v1:"
	// secretInfo is the HKDF info string. HKDF is one way so this secret cannot be
	// walked back to the signing key, the two stay independent.
	secretInfo = "monetr-pow-secret-v1"
	// skewLeeway is the clock skew we tolerate on expiry, also padded onto the
	// marker TTL.
	skewLeeway = 30 * time.Second

	// The token is a fixed length binary blob, laid out as:
	//   [1]  version
	//   [1]  purpose
	//   [16] random
	//   [8]  issued at, unix seconds, big-endian
	//   [2]  difficulty, big-endian
	//   [32] HMAC-SHA256 of the preceding 28 bytes
	// then base32 (no padding) encoded for transport.
	randomLength = 16
	signedLength = 28 // Everything except the trailing HMAC.
	tokenLength  = 60

	versionOffset    = 0
	purposeOffset    = 1
	randomOffset     = 2
	issuedAtOffset   = 18
	difficultyOffset = 26
	macOffset        = 28
)

// tokenEncoding stringifies the binary token. base32 keeps it on an unambiguous,
// copy-paste friendly alphabet, no +/ or -_ confusion between base64 variants.
var tokenEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

type Challenge struct {
	Token      string `json:"challenge"`
	Difficulty int    `json:"difficulty"`
}

// Challenger issues and verifies server signed proof of work challenges. The
// whole point is to make the unauthenticated authentication endpoints expensive
// to hit in an automated way without making real users wait. See the
// [config.ProofOfWork] documentation for the bigger picture.
type Challenger interface {
	// Issue will create a new challenge bound to the given purpose. It also
	// records a single-use marker in the cache so the resulting challenge can
	// only be redeemed once.
	Issue(ctx context.Context, purpose Purpose) (*Challenge, error)
	// Verify will check that the provided token and nonce are a valid solution
	// for the given purpose. It returns one of the typed errors above when the
	// challenge is not acceptable, or nil when everything checks out. A
	// successful verification consumes the challenge so it cannot be used again.
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

// DeriveSecret derives the 32 byte HMAC secret from the existing ED25519 seed
// via HKDF-SHA256. HKDF is one way, so this stays independent of the signing key
// and we do not have to manage a separate secret just for proof of work.
func DeriveSecret(seed []byte) []byte {
	secret, err := hkdf.Key(sha256.New, seed, nil, secretInfo, 32)
	if err != nil {
		// Only errors when the length is too large for the hash, which 32 bytes
		// out of SHA-256 never is.
		panic(errors.Wrap(err, "failed to derive proof of work secret"))
	}
	return secret
}

func (c *challengerBase) Issue(ctx context.Context, purpose Purpose) (*Challenge, error) {
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

	// Everything except the trailing HMAC is signed. Signing the purpose locks the
	// token to one endpoint, signing the difficulty stops anyone grinding cheap
	// low difficulty tokens and reusing them after we raise the bar.
	token := make([]byte, tokenLength)
	token[versionOffset] = tokenVersion
	token[purposeOffset] = byte(purpose)
	copy(token[randomOffset:randomOffset+randomLength], random)
	binary.BigEndian.PutUint64(token[issuedAtOffset:issuedAtOffset+8], uint64(issuedAt.Unix()))
	binary.BigEndian.PutUint16(token[difficultyOffset:difficultyOffset+2], uint16(c.difficulty))
	copy(token[macOffset:], c.sign(token[:signedLength]))

	encoded := tokenEncoding.EncodeToString(token)

	// Plant the single-use marker, flipped to "consumed" in Verify. The key and
	// values are HMAC-derived so a redis-write-only attacker without the secret
	// cannot plant one. The TTL runs a bit past the lifetime so a still-valid
	// challenge always has a marker to flip.
	// NOTE The marker TTL is on the real wall clock (the cache uses time.Now), not
	// c.clock. That is fine, the c.clock expiry check in Verify is the real gate,
	// the marker is only for replay and cleanup.
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
	}, nil
}

func (c *challengerBase) Verify(ctx context.Context, expectedPurpose Purpose, token string, nonce uint64) (returnErr error) {
	span := crumbs.StartFnTrace(ctx)
	defer span.Finish()

	// Record a metric for every verification, labeled by outcome.
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

	// Verify the HMAC before trusting any other field. Constant time so we do not
	// leak the signature through timing.
	if subtle.ConstantTimeCompare(raw[macOffset:], c.sign(raw[:signedLength])) != 1 {
		span.Status = sentry.SpanStatusInvalidArgument
		result = "invalid"
		return errors.WithStack(ErrInvalidChallenge)
	}

	// Purpose (now trusted) must match the endpoint, keeps a register challenge
	// from being redeemed on login.
	if Purpose(raw[purposeOffset]) != expectedPurpose {
		span.Status = sentry.SpanStatusInvalidArgument
		result = "invalid"
		return errors.WithStack(ErrInvalidChallenge)
	}

	// Reject tokens minted below our current difficulty policy.
	difficulty := int(binary.BigEndian.Uint16(raw[difficultyOffset : difficultyOffset+2]))
	if difficulty < c.difficulty {
		span.Status = sentry.SpanStatusInvalidArgument
		result = "invalid"
		return errors.WithStack(ErrInvalidChallenge)
	}

	// The leeway tolerates a token that looks issued slightly in the future
	// because the issuing instance's clock is a touch ahead.
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

	// Replay protection: atomically flip the marker unused -> consumed. No swap
	// means already used or gone, either way reject. The random comes from the
	// trusted token so we rederive the same key and values Issue wrote.
	random := raw[randomOffset : randomOffset+randomLength]
	swapped, err := c.cache.CompareAndSwap(
		span.Context(),
		c.replayKey(random),
		c.unusedValue(random),
		c.consumedValue(random),
		c.lifetime+skewLeeway,
	)
	if err != nil {
		// Fail closed, better to reject than risk a replay when redis is down.
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

// recordVerify emits the verification metrics. stats is nil in many test setups.
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

// replayKey is the cache key for a challenge's marker, HMAC'd so it is not
// predictable to someone who can write redis but does not know our secret.
func (c *challengerBase) replayKey(random []byte) string {
	return "pow:" + tokenEncoding.EncodeToString(c.tag("pow-key:", random))
}

func (c *challengerBase) unusedValue(random []byte) []byte {
	return c.tag("pow-unused:", random)
}

func (c *challengerBase) consumedValue(random []byte) []byte {
	return c.tag("pow-consumed:", random)
}

// tag HMACs a label with the random bytes. The distinct labels keep the key and
// the two marker values different even though they share the same random.
func (c *challengerBase) tag(label string, random []byte) []byte {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(label))
	mac.Write(random)
	return mac.Sum(nil)
}

// computeProofDigest is the exact hash the client solves: prefix, token, a colon,
// then the nonce as 8 big-endian bytes. MUST match the frontend solver in
// interface/src/util/proofOfWorkSolver.ts, the shared test vectors guard drift.
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

// leadingZeroBits counts leading zero bits (not bytes or hex digits) so the
// difficulty can be tuned one bit at a time.
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
